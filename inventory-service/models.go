package main

import "time"

type Product struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	PricePerLiter float64 `json:"price_per_liter"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type InventoryStock struct {
	ID             int     `json:"id"`
	ProductID      int     `json:"product_id"`
	TimeSlot       string  `json:"time_slot"`
	TotalStock     float64 `json:"total_stock"`
	BookedStock    float64 `json:"booked_stock"`
	AvailableStock float64 `json:"available_stock"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type StockHistory struct {
	ID            int       `json:"id"`
	ProductID     int       `json:"product_id"`
	TimeSlot      string    `json:"time_slot"`
	ChangeType    string    `json:"change_type"`
	Quantity      float64   `json:"quantity"`
	PreviousValue float64   `json:"previous_value"`
	NewValue      float64   `json:"new_value"`
	Reason        string    `json:"reason"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
}

type NotificationRequest struct {
	ID           int       `json:"id"`
	CustomerName string    `json:"customer_name"`
	PhoneNumber  string    `json:"phone_number"`
	ProductID    int       `json:"product_id"`
	TimeSlot     string    `json:"time_slot"`
	Quantity     float64   `json:"quantity"`
	Status       string    `json:"status"`
	RequestDate  string    `json:"request_date"`
	RequestTime  string    `json:"request_time"`
	NotifiedAt   *time.Time `json:"notified_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type StockAdjustment struct {
	ProductID int     `json:"product_id"`
	TimeSlot  string  `json:"time_slot"`
	Quantity  float64 `json:"quantity"`
	Reason    string  `json:"reason"`
}

type BookingRequest struct {
	ProductID int     `json:"product_id"`
	TimeSlot  string  `json:"time_slot"`
	Quantity  float64 `json:"quantity"`
	Reason    string  `json:"reason"`
}

type PriceUpdate struct {
	Price float64 `json:"price"`
}

type AnalyticsSummary struct {
	TotalStock     float64 `json:"total_stock"`
	TotalBooked    float64 `json:"total_booked"`
	TotalAvailable float64 `json:"total_available"`
	DailyRevenue   float64 `json:"daily_revenue"`
	Products       []ProductSummary `json:"products"`
}

type ProductSummary struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	Type        string  `json:"type"`
	Price       float64 `json:"price"`
	Morning     StockSummary `json:"morning"`
	Evening     StockSummary `json:"evening"`
}

type StockSummary struct {
	TotalStock     float64 `json:"total_stock"`
	BookedStock    float64 `json:"booked_stock"`
	AvailableStock float64 `json:"available_stock"`
	Revenue        float64 `json:"revenue"`
}