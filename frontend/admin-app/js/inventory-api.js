// Inventory API Client
class InventoryAPI {
    constructor() {
        this.baseURL = '/api/v1/inventory';
    }

    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        };

        try {
            const response = await fetch(url, config);
            const data = await response.json();
            
            if (!response.ok) {
                // Extract error message from response
                const errorMessage = data.error || `HTTP error! status: ${response.status}`;
                throw new Error(errorMessage);
            }
            return data;
        } catch (error) {
            console.error('API request failed:', error);
            throw error;
        }
    }

    // Products
    async getProducts() {
        return this.request('/products');
    }

    async updateProductPrice(productId, price) {
        return this.request(`/products/${productId}/price`, {
            method: 'PUT',
            body: JSON.stringify({ price })
        });
    }

    // Stock Management
    async getStock() {
        return this.request('/stock');
    }

    async updateStock(productId, timeSlot, totalStock) {
        return this.request(`/stock/${productId}/${timeSlot}`, {
            method: 'PUT',
            body: JSON.stringify({ total_stock: totalStock })
        });
    }

    async adjustStock(productId, timeSlot, quantity, reason = 'Manual adjustment') {
        return this.request(`/stock/${productId}/${timeSlot}/adjust`, {
            method: 'POST',
            body: JSON.stringify({ 
                product_id: productId,
                time_slot: timeSlot,
                quantity,
                reason 
            })
        });
    }

    // Bookings
    async addBooking(productId, timeSlot, quantity, reason = 'Subscription booking') {
        return this.request('/bookings', {
            method: 'POST',
            body: JSON.stringify({
                product_id: productId,
                time_slot: timeSlot,
                quantity,
                reason
            })
        });
    }

    async removeBooking(productId, timeSlot, quantity, reason = 'Subscription cancellation') {
        return this.request('/bookings', {
            method: 'DELETE',
            body: JSON.stringify({
                product_id: productId,
                time_slot: timeSlot,
                quantity,
                reason
            })
        });
    }

    // Notifications
    async getNotifications(status = '') {
        const query = status ? `?status=${status}` : '';
        return this.request(`/notifications${query}`);
    }

    async createNotification(customerName, phoneNumber, productId, timeSlot, quantity) {
        return this.request('/notifications', {
            method: 'POST',
            body: JSON.stringify({
                customer_name: customerName,
                phone_number: phoneNumber,
                product_id: productId,
                time_slot: timeSlot,
                quantity
            })
        });
    }

    async updateNotificationStatus(id, status) {
        return this.request(`/notifications/${id}/status`, {
            method: 'PUT',
            body: JSON.stringify({ status })
        });
    }

    async deleteNotification(id) {
        return this.request(`/notifications/${id}`, {
            method: 'DELETE'
        });
    }

    // Analytics
    async getAnalyticsSummary() {
        return this.request('/analytics/summary');
    }
}

// Global instance
window.inventoryAPI = new InventoryAPI();