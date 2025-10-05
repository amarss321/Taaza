// Address API integration
class AddressAPI {
    constructor() {
        this.baseURL = '/api/v1/users';
    }

    // Get auth token from localStorage
    getAuthToken() {
        return localStorage.getItem('authToken') || localStorage.getItem('token');
    }

    // Get auth headers
    getAuthHeaders() {
        const token = this.getAuthToken();
        return {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        };
    }

    // Get API URL with fallback
    getApiUrl(endpoint) {
        // Try API gateway first, then direct service
        return `/api/v1/users${endpoint}`;
    }

    // Get all addresses
    async getAddresses() {
        try {
            const response = await fetch(this.getApiUrl('/addresses'), {
                method: 'GET',
                headers: this.getAuthHeaders()
            });

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(`HTTP ${response.status}: ${errorText}`);
            }

            const data = await response.json();
            return data.addresses || [];
        } catch (error) {
            console.error('Error fetching addresses:', error);
            throw error;
        }
    }

    // Create new address
    async createAddress(addressData) {
        try {
            // Validate required fields
            const required = ['label', 'address_line', 'city', 'zip_code', 'country'];
            for (const field of required) {
                if (!addressData[field] || addressData[field].trim() === '') {
                    throw new Error(`${field} is required`);
                }
            }

            const response = await fetch(this.getApiUrl('/addresses'), {
                method: 'POST',
                headers: this.getAuthHeaders(),
                body: JSON.stringify(addressData)
            });

            if (!response.ok) {
                const errorText = await response.text();
                let errorData;
                try {
                    errorData = JSON.parse(errorText);
                } catch {
                    errorData = { error: errorText };
                }
                throw new Error(errorData.error || `HTTP ${response.status}: ${errorText}`);
            }

            return await response.json();
        } catch (error) {
            console.error('Error creating address:', error);
            throw error;
        }
    }

    // Update address
    async updateAddress(addressId, addressData) {
        try {
            const response = await fetch(`${this.baseURL}/addresses/${addressId}`, {
                method: 'PUT',
                headers: this.getAuthHeaders(),
                body: JSON.stringify(addressData)
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
            }

            return await response.json();
        } catch (error) {
            console.error('Error updating address:', error);
            throw error;
        }
    }

    // Delete address
    async deleteAddress(addressId) {
        try {
            const response = await fetch(`${this.baseURL}/addresses/${addressId}`, {
                method: 'DELETE',
                headers: this.getAuthHeaders()
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
            }

            return await response.json();
        } catch (error) {
            console.error('Error deleting address:', error);
            throw error;
        }
    }

    // Set default address
    async setDefaultAddress(addressId) {
        try {
            const response = await fetch(`${this.baseURL}/addresses/${addressId}/default`, {
                method: 'PUT',
                headers: this.getAuthHeaders()
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
            }

            return await response.json();
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