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
        }
        return token;
    }

    // Get cookie value
    getCookie(name) {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; ${name}=`);
        if (parts.length === 2) return parts.pop().split(';').shift();
    }

    // Redirect to login if no token
    redirectToLogin() {
        if (!window.location.href.includes('login')) {
            window.location.href = '/auth/3-Taaza-Login.html';
        }
    }

    // Enhanced API call with better error handling
    async makeAPIRequest(url, options = {}) {
        const token = this.getAuthToken();
        
        if (!token) {
            throw new Error('No authentication token found. Please login first.');
        }
        
        const defaultHeaders = {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        };

        try {
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
                throw new Error(`HTTP ${response.status}: ${errorText}`);
            }

            return await response.json();
        } catch (error) {
            console.error('API Request failed:', error);
            throw error;
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

    // Migrate localStorage addresses to backend
    async migrateLocalStorageAddresses() {
        try {
            const localAddresses = JSON.parse(localStorage.getItem('addresses') || '[]');
            const defaultIndex = parseInt(localStorage.getItem('defaultAddress') || '0');

            if (localAddresses.length === 0) {
                return { migrated: 0 };
            }

            let migratedCount = 0;
            for (let i = 0; i < localAddresses.length; i++) {
                const addr = localAddresses[i];
                const addressData = {
                    label: addr.label,
                    address_line: addr.line,
                    city: addr.city,
                    state: addr.state || '',
                    zip_code: addr.zip,
                    country: addr.country,
                    latitude: addr.lat || null,
                    longitude: addr.lng || null,
                    is_default: i === defaultIndex
                };

                try {
                    await this.createAddress(addressData);
                    migratedCount++;
                } catch (error) {
                    console.error(`Failed to migrate address ${i}:`, error);
                }
            }

            // Clear localStorage after successful migration
            if (migratedCount > 0) {
                localStorage.removeItem('addresses');
                localStorage.removeItem('defaultAddress');
            }

            return { migrated: migratedCount };
        } catch (error) {
            console.error('Error migrating addresses:', error);
            throw error;
        }
    }
}

// Export for use in addresses.html
window.AddressAPI = AddressAPI;