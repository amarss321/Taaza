// Database-based session management utility
class DBSessionManager {
    constructor() {
        this.timeoutDuration = 7 * 24 * 60 * 60 * 1000; // 7 days
        this.warningDuration = 60 * 60 * 1000; // 1 hour before timeout
        this.timeoutId = null;
        this.warningId = null;
        this.lastActivity = Date.now();
        
        this.init();
    }
    
    init() {
        // Track user activity
        const events = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart', 'click'];
        events.forEach(event => {
            document.addEventListener(event, () => this.resetTimer(), true);
        });
        
        // Start session timer
        this.resetTimer();
        
        // Check for existing session on page load
        this.checkSession();
    }
    
    resetTimer() {
        this.lastActivity = Date.now();
        
        // Clear existing timers
        if (this.timeoutId) clearTimeout(this.timeoutId);
        if (this.warningId) clearTimeout(this.warningId);
        
        // Set warning timer
        this.warningId = setTimeout(() => this.showWarning(), this.timeoutDuration - this.warningDuration);
        
        // Set logout timer
        this.timeoutId = setTimeout(() => this.logout(), this.timeoutDuration);
    }
    
    showWarning() {
        const modal = document.createElement('div');
        modal.id = 'sessionWarning';
        modal.className = 'fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50';
        modal.innerHTML = `
            <div class="bg-white rounded-lg p-6 max-w-sm mx-4">
                <h3 class="text-lg font-bold text-gray-900 mb-2">Session Expiring</h3>
                <p class="text-gray-600 mb-4">Your session will expire in 1 hour due to inactivity.</p>
                <div class="flex gap-2">
                    <button id="extendSession" class="flex-1 bg-primary text-background-dark px-4 py-2 rounded font-medium">
                        Stay Logged In
                    </button>
                    <button id="logoutNow" class="flex-1 bg-gray-300 text-gray-700 px-4 py-2 rounded font-medium">
                        Logout
                    </button>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        document.getElementById('extendSession').onclick = () => {
            this.resetTimer();
            document.body.removeChild(modal);
        };
        
        document.getElementById('logoutNow').onclick = () => {
            this.logout();
        };
    }
    
    async logout() {
        try {
            // Call logout API to invalidate session on server
            await fetch('/api/v1/users/logout', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.getToken()}`,
                    'Content-Type': 'application/json'
                }
            });
        } catch (error) {
            console.error('Logout API error:', error);
        }
        
        // Clear any remaining client-side data
        this.clearSession();
        
        // Show logout message
        alert('Your session has expired. Please login again.');
        
        // Redirect to login
        window.location.href = '/auth/3-Taaza-Login.html';
    }
    
    clearSession() {
        // Remove any remaining localStorage items (for migration period)
        const keysToRemove = [
            'authToken', 'userName', 'userEmail', 'userMobile',
            'userSubscriptions', 'adminStockData', 'deliverySchedule',
            'milkSubscription', 'morningDelivery', 'eveningDelivery',
            'morningMilkType', 'morningQuantity', 'morningFrequency',
            'morningTimeSlot', 'morningDays', 'eveningMilkType',
            'eveningQuantity', 'eveningFrequency', 'eveningTimeSlot',
            'eveningDays', 'subscriptionAddress', 'pendingSubscription'
        ];
        
        keysToRemove.forEach(key => localStorage.removeItem(key));
    }
    
    getToken() {
        // During migration, check localStorage first, then cookies
        return localStorage.getItem('authToken') || this.getCookie('authToken');
    }
    
    getCookie(name) {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; ${name}=`);
        if (parts.length === 2) return parts.pop().split(';').shift();
        return null;
    }
    
    async checkSession() {
        const token = this.getToken();
        if (!token && this.requiresAuth()) {
            this.logout();
            return;
        }
        
        // If we have a token, validate it with backend
        if (token && this.requiresAuth()) {
            try {
                const response = await fetch('/api/v1/users/profile', {
                    headers: {
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json'
                    }
                });
                
                if (!response.ok) {
                    // Token is invalid/expired on backend
                    this.logout();
                }
            } catch (error) {
                console.error('Session validation failed:', error);
                // On network error, don't logout - might be temporary
            }
        }
    }
    
    requiresAuth() {
        const publicPages = [
            '3-Taaza-Login.html',
            '4-OTP-Verification.html', 
            '5-User-Registration.html',
            'index.html',
            '2-onboarding.html'
        ];
        
        const currentPage = window.location.pathname.split('/').pop();
        return !publicPages.includes(currentPage);
    }
}

// Initialize session manager
if (typeof window !== 'undefined') {
    window.dbSessionManager = new DBSessionManager();
}