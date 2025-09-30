// Enhanced Address API with better error handling
class AddressAPI {
    constructor() {
        this.baseURL = '/api/v1/users';
    }

    // Enhanced token detection across multiple storage methods
    getAuthToken() {
        const token = 
            localStorage.getItem('authToken') ||
            sessionStorage.getItem('authToken') ||
            this.getCookie('authToken') ||
            localStorage.getItem('token');
        
        if (!token) {
            console.warn('No auth token found in any storage');
            return null;
        }
        return token;
    }

    // Get cookie value
    getCookie(name) {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; ${name}=`);
        if (parts.length === 2) return parts.pop().split(';').shift();
        return null;
    }

    // Get auth headers with device ID
    getAuthHeaders() {
        const token = this.getAuthToken();
        const deviceId = this.getDeviceId();
        
        const headers = {
            'Content-Type': 'application/json'
        };
        
        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }
        
        if (deviceId) {
            headers['X-Device-ID'] = deviceId;
        }
        
        return headers;
    }

    // Get or create device ID
    getDeviceId() {
        let deviceId = localStorage.getItem('deviceId');
        if (!deviceId) {
            deviceId = 'device_' + Math.random().toString(36).substr(2, 9) + '_' + Date.now();
            localStorage.setItem('deviceId', deviceId);
        }
        return deviceId;
    }

    // Enhanced API call with better error handling
    async makeAPIRequest(url, options = {}) {
        const token = this.getAuthToken();
        
        if (!token) {
            throw new Error('No authentication token found. Please login first.');
        }
        
        const defaultHeaders = this.getAuthHeaders();

        try {
            console.log('Making API request to:', url, 'with headers:', defaultHeaders);
            
            const response = await fetch(url, {
                ...options,
                headers: { ...defaultHeaders, ...options.headers }
            });

            // Handle unauthorized (401) - redirect to login
            if (response.status === 401) {
                console.warn('Token expired, redirecting to login');
                this.redirectToLogin();
                throw new Error('Authentication required');
            }

            if (!response.ok) {
                const errorText = await response.text();
                console.error('API Error Response:', errorText);
                throw new Error(`HTTP ${response.status}: ${errorText}`);
            }

            const data = await response.json();
            console.log('API Success Response:', data);
            return data;
        } catch (error) {
            console.error('API Request failed:', error);
            throw error;
        }
    }

    // Redirect to login if no token
    redirectToLogin() {
        if (!window.location.href.includes('login')) {
            window.location.href = '/auth/3-Taaza-Login.html';
        }
    }

    // Get all addresses with enhanced error handling
    async getAddresses() {
        try {
            const data = await this.makeAPIRequest(`${this.baseURL}/addresses`, {
                method: 'GET'
            });
            return data.addresses || [];
        } catch (error) {
            console.error('Error fetching addresses:', error);
            throw error;
        }
    }

    // Create new address with validation
    async createAddress(addressData) {
        try {
            console.log('Creating address with data:', addressData);
            console.log('Auth token available:', !!this.getAuthToken());
            console.log('Device ID:', this.getDeviceId());
            
            // Validate required fields
            const required = ['label', 'address_line', 'city', 'zip_code', 'country'];
            for (const field of required) {
                if (!addressData[field]?.trim()) {
                    throw new Error(`Missing required field: ${field}`);
                }
            }

            const result = await this.makeAPIRequest(`${this.baseURL}/addresses`, {
                method: 'POST',
                body: JSON.stringify(addressData)
            });
            
            console.log('Address created successfully:', result);
            return result;
        } catch (error) {
            console.error('Error creating address:', error);
            throw error;
        }
    }

    // Update existing address
    async updateAddress(addressId, addressData) {
        try {
            const result = await this.makeAPIRequest(`${this.baseURL}/addresses/${addressId}`, {
                method: 'PUT',
                body: JSON.stringify(addressData)
            });
            return result;
        } catch (error) {
            console.error('Error updating address:', error);
            throw error;
        }
    }

    // Delete address
    async deleteAddress(addressId) {
        try {
            const result = await this.makeAPIRequest(`${this.baseURL}/addresses/${addressId}`, {
                method: 'DELETE'
            });
            return result;
        } catch (error) {
            console.error('Error deleting address:', error);
            throw error;
        }
    }

    // Set default address
    async setDefaultAddress(addressId) {
        try {
            const result = await this.makeAPIRequest(`${this.baseURL}/addresses/${addressId}/default`, {
                method: 'PUT'
            });
            return result;
        } catch (error) {
            console.error('Error setting default address:', error);
            throw error;
        }
    }

    // Test authentication
    async testAuth() {
        try {
            const response = await fetch(`${this.baseURL}/profile`, {
                headers: this.getAuthHeaders()
            });
            return response.ok;
        } catch (error) {
            return false;
        }
    }
}

// Export for use in addresses.html
window.AddressAPI = AddressAPI;