// Database API utility to replace localStorage operations
class DataAPI {
    constructor() {
        this.baseURL = '/api/v1/users';
    }

    getAuthHeaders() {
        const token = localStorage.getItem('authToken') || this.getCookie('authToken');
        return {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        };
    }

    getCookie(name) {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; ${name}=`);
        if (parts.length === 2) return parts.pop().split(';').shift();
        return null;
    }

    async handleResponse(response) {
        if (!response.ok) {
            if (response.status === 401) {
                // Token expired, redirect to login
                window.location.href = '/auth/3-Taaza-Login.html';
                return null;
            }
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        return response.json();
    }

    // Subscription methods
    async getSubscriptions() {
        try {
            const response = await fetch(`${this.baseURL}/subscriptions`, {
                headers: this.getAuthHeaders()
            });
            return await this.handleResponse(response);
        } catch (error) {
            console.error('Error fetching subscriptions:', error);
            return [];
        }
    }

    async createSubscription(subscriptionData) {
        try {
            const response = await fetch(`${this.baseURL}/subscriptions`, {
                method: 'POST',
                headers: this.getAuthHeaders(),
                body: JSON.stringify(subscriptionData)
            });
            return await this.handleResponse(response);
        } catch (error) {
            console.error('Error creating subscription:', error);
            throw error;
        }
    }

    async updateSubscription(subscriptionId, subscriptionData) {
        try {
            const response = await fetch(`${this.baseURL}/subscriptions/${subscriptionId}`, {
                method: 'PUT',
                headers: this.getAuthHeaders(),
                body: JSON.stringify(subscriptionData)
            });
            return await this.handleResponse(response);
        } catch (error) {
            console.error('Error updating subscription:', error);
            throw error;
        }
    }

    async deleteSubscription(subscriptionId) {
        try {
            const response = await fetch(`${this.baseURL}/subscriptions/${subscriptionId}`, {
                method: 'DELETE',
                headers: this.getAuthHeaders()
            });
            return await this.handleResponse(response);
        } catch (error) {
            console.error('Error deleting subscription:', error);
            throw error;
        }
    }

    // Preference methods
    async getPreferences() {
        try {
            const response = await fetch(`${this.baseURL}/preferences`, {
                headers: this.getAuthHeaders()
            });
            return await this.handleResponse(response);
        } catch (error) {
            console.error('Error fetching preferences:', error);
            return {};
        }
    }

    async setPreference(key, value) {
        try {
            const response = await fetch(`${this.baseURL}/preferences`, {
                method: 'POST',
                headers: this.getAuthHeaders(),
                body: JSON.stringify({ key, value })
            });
            return await this.handleResponse(response);
        } catch (error) {
            console.error('Error setting preference:', error);
            throw error;
        }
    }

    async setPreferences(preferences) {
        try {
            const response = await fetch(`${this.baseURL}/preferences`, {
                method: 'PUT',
                headers: this.getAuthHeaders(),
                body: JSON.stringify(preferences)
            });
            return await this.handleResponse(response);
        } catch (error) {
            console.error('Error setting preferences:', error);
            throw error;
        }
    }

    // Helper methods for common localStorage replacements
    async getItem(key) {
        const preferences = await this.getPreferences();
        return preferences[key] || null;
    }

    async setItem(key, value) {
        return await this.setPreference(key, typeof value === 'string' ? value : JSON.stringify(value));
    }

    async removeItem(key) {
        return await this.setPreference(key, '');
    }

    // Migration helper - move localStorage data to database
    async migrateLocalStorageData() {
        const token = localStorage.getItem('authToken');
        if (!token) return;

        try {
            // Migrate user preferences
            const preferencesToMigrate = {
                'adminActiveSection': localStorage.getItem('adminActiveSection'),
                'adminActiveMilkTab': localStorage.getItem('adminActiveMilkTab'),
                'adminFullscreenMode': localStorage.getItem('adminFullscreenMode'),
                'editingSubscription': localStorage.getItem('editingSubscription'),
                'subscriptionUpdate': localStorage.getItem('subscriptionUpdate'),
                'milkSubscription': localStorage.getItem('milkSubscription')
            };

            // Filter out null values
            const validPreferences = {};
            Object.keys(preferencesToMigrate).forEach(key => {
                if (preferencesToMigrate[key] !== null) {
                    validPreferences[key] = preferencesToMigrate[key];
                }
            });

            if (Object.keys(validPreferences).length > 0) {
                await this.setPreferences(validPreferences);
            }

            // Migrate subscription data if exists
            const morningEnabled = localStorage.getItem('morningDelivery') === 'true';
            const eveningEnabled = localStorage.getItem('eveningDelivery') === 'true';

            if (morningEnabled || eveningEnabled) {
                const subscriptionData = {
                    subscription_type: 'milk',
                    morning_enabled: morningEnabled,
                    morning_milk_type: localStorage.getItem('morningMilkType'),
                    morning_quantity: parseFloat(localStorage.getItem('morningQuantity') || '0'),
                    morning_frequency: localStorage.getItem('morningFrequency'),
                    morning_time_slot: localStorage.getItem('morningTimeSlot'),
                    morning_days: JSON.parse(localStorage.getItem('morningDays') || '[]'),
                    evening_enabled: eveningEnabled,
                    evening_milk_type: localStorage.getItem('eveningMilkType'),
                    evening_quantity: parseFloat(localStorage.getItem('eveningQuantity') || '0'),
                    evening_frequency: localStorage.getItem('eveningFrequency'),
                    evening_time_slot: localStorage.getItem('eveningTimeSlot'),
                    evening_days: JSON.parse(localStorage.getItem('eveningDays') || '[]'),
                    address_data: JSON.parse(localStorage.getItem('subscriptionAddress') || '{}'),
                    status: 'active'
                };

                await this.createSubscription(subscriptionData);
            }

            console.log('LocalStorage data migrated successfully');
        } catch (error) {
            console.error('Migration failed:', error);
        }
    }
}

// Initialize global instance
if (typeof window !== 'undefined') {
    window.dataAPI = new DataAPI();
    
    // Auto-migrate on first load if user is authenticated
    document.addEventListener('DOMContentLoaded', () => {
        const token = localStorage.getItem('authToken');
        if (token) {
            window.dataAPI.migrateLocalStorageData();
        }
    });
}