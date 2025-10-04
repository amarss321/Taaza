// Get milk prices from API with cache-busting for real-time updates
async function getMilkPricesFromAPI() {
    try {
        const response = await fetch('/api/v1/inventory/products?' + Date.now());
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        
        const result = await response.json();
        console.log('Fresh prices loaded:', result);
        
        const products = result.products || [];
        const prices = {
            buffalo: products.find(p => p.type === 'buffalo')?.price_per_liter || 80,
            cow: products.find(p => p.type === 'cow')?.price_per_liter || 70
        };
        
        console.log('Current prices:', prices);
        return prices;
    } catch (error) {
        console.error('Price fetch failed:', error);
        return { buffalo: 80, cow: 70 };
    }
}