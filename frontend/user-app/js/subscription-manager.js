// Subscription manager using database API instead of localStorage
class SubscriptionManager {
    constructor() {
        this.dataAPI = window.dataAPI;
    }

    // Get all user subscriptions
    async getSubscriptions() {
        try {
            return await this.dataAPI.getSubscriptions();
        } catch (error) {
            console.error('Error fetching subscriptions:', error);
            return [];
        }
    }

    // Create a new subscription
    async createSubscription(subscriptionData) {
        try {
            const subscription = {
                subscription_type: 'milk',
                morning_enabled: subscriptionData.morningEnabled || false,
                morning_milk_type: subscriptionData.morningMilkType || null,
                morning_quantity: subscriptionData.morningQuantity || 0,
                morning_frequency: subscriptionData.morningFrequency || null,
                morning_time_slot: subscriptionData.morningTimeSlot || null,
                morning_days: subscriptionData.morningDays || [],
                evening_enabled: subscriptionData.eveningEnabled || false,
                evening_milk_type: subscriptionData.eveningMilkType || null,
                evening_quantity: subscriptionData.eveningQuantity || 0,
                evening_frequency: subscriptionData.eveningFrequency || null,
                evening_time_slot: subscriptionData.eveningTimeSlot || null,
                evening_days: subscriptionData.eveningDays || [],
                address_data: subscriptionData.addressData || {},
                status: 'active'
            };

            return await this.dataAPI.createSubscription(subscription);
        } catch (error) {
            console.error('Error creating subscription:', error);
            throw error;
        }
    }

    // Update existing subscription
    async updateSubscription(subscriptionId, subscriptionData) {
        try {
            return await this.dataAPI.updateSubscription(subscriptionId, subscriptionData);
        } catch (error) {
            console.error('Error updating subscription:', error);
            throw error;
        }
    }

    // Cancel subscription
    async cancelSubscription(subscriptionId) {
        try {
            return await this.dataAPI.deleteSubscription(subscriptionId);
        } catch (error) {
            console.error('Error cancelling subscription:', error);
            throw error;
        }
    }

    // Helper methods for common operations
    async hasActiveSubscription() {
        const subscriptions = await this.getSubscriptions();
        return subscriptions.length > 0;
    }

    async getMorningSubscription() {
        const subscriptions = await this.getSubscriptions();
        return subscriptions.find(sub => sub.morning_enabled);
    }

    async getEveningSubscription() {
        const subscriptions = await this.getSubscriptions();
        return subscriptions.find(sub => sub.evening_enabled);
    }

    // Migration helper - convert localStorage data to subscription
    async migrateFromLocalStorage() {
        const morningEnabled = localStorage.getItem('morningDelivery') === 'true';
        const eveningEnabled = localStorage.getItem('eveningDelivery') === 'true';

        if (!morningEnabled && !eveningEnabled) {
            return null;
        }

        const subscriptionData = {
            morningEnabled,
            morningMilkType: localStorage.getItem('morningMilkType'),
            morningQuantity: parseFloat(localStorage.getItem('morningQuantity') || '0'),
            morningFrequency: localStorage.getItem('morningFrequency'),
            morningTimeSlot: localStorage.getItem('morningTimeSlot'),
            morningDays: JSON.parse(localStorage.getItem('morningDays') || '[]'),
            eveningEnabled,
            eveningMilkType: localStorage.getItem('eveningMilkType'),
            eveningQuantity: parseFloat(localStorage.getItem('eveningQuantity') || '0'),
            eveningFrequency: localStorage.getItem('eveningFrequency'),
            eveningTimeSlot: localStorage.getItem('eveningTimeSlot'),
            eveningDays: JSON.parse(localStorage.getItem('eveningDays') || '[]'),
            addressData: JSON.parse(localStorage.getItem('subscriptionAddress') || '{}')
        };

        try {
            const result = await this.createSubscription(subscriptionData);
            console.log('Subscription migrated from localStorage');
            return result;
        } catch (error) {
            console.error('Failed to migrate subscription:', error);
            return null;
        }
    }
}

// Initialize global instance
if (typeof window !== 'undefined') {
    window.subscriptionManager = new SubscriptionManager();
}