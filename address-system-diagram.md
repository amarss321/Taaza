# Taaza Address System - Data Flow & Storage Diagram

## System Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              FRONTEND LAYER                                     │
├─────────────────────────────────────────────────────────────────────────────────┤
│  📱 addresses.html                                                              │
│  ├── User Interface Components:                                                 │
│  │   ├── Address List Display                                                  │
│  │   ├── Add/Edit Address Modal                                                │
│  │   ├── Map Integration (Leaflet)                                             │
│  │   └── Default Address Management                                            │
│  │                                                                             │
│  📜 addresses-api.js                                                            │
│  ├── API Communication Layer:                                                  │
│  │   ├── Authentication Token Management                                       │
│  │   ├── HTTP Request Handling                                                 │
│  │   ├── Error Handling & Redirects                                           │
│  │   └── LocalStorage Migration                                               │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        │ HTTP Requests
                                        │ (Bearer Token Auth)
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              API GATEWAY                                        │
├─────────────────────────────────────────────────────────────────────────────────┤
│  🌐 nginx (Port 80)                                                             │
│  ├── Route: /api/v1/users/addresses/*                                          │
│  ├── Load Balancing                                                            │
│  ├── SSL Termination                                                           │
│  └── Request Forwarding                                                        │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        │ Proxied Requests
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            USER SERVICE LAYER                                   │
├─────────────────────────────────────────────────────────────────────────────────┤
│  🔧 user-service (Go/Gin)                                                      │
│  │                                                                             │
│  ├── 🔐 Middleware Stack:                                                      │
│  │   ├── AuthMiddleware() - JWT Token Validation                              │
│  │   ├── SecurityHeaders()                                                    │
│  │   ├── RateLimit()                                                          │
│  │   ├── ValidateInput()                                                      │
│  │   └── SanitizeInput()                                                      │
│  │                                                                             │
│  ├── 📍 Address Handlers (handlers/addresses.go):                             │
│  │   ├── GET    /addresses        → GetAddresses()                           │
│  │   ├── POST   /addresses        → CreateAddress()                          │
│  │   ├── PUT    /addresses/:id    → UpdateAddress()                          │
│  │   ├── DELETE /addresses/:id    → DeleteAddress()                          │
│  │   └── PUT    /addresses/:id/default → SetDefaultAddress()                 │
│  │                                                                             │
│  └── 📊 Address Model:                                                         │
│      ├── ID, UserID, Label                                                    │
│      ├── AddressLine, City, State                                             │
│      ├── ZipCode, Country                                                     │
│      ├── Latitude, Longitude (GPS)                                            │
│      ├── IsDefault (Boolean)                                                  │
│      └── CreatedAt, UpdatedAt                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        │ SQL Queries
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            DATABASE LAYER                                       │
├─────────────────────────────────────────────────────────────────────────────────┤
│  🗄️  PostgreSQL Database                                                       │
│  │                                                                             │
│  ├── 📋 Database: taaza_users                                                  │
│  │                                                                             │
│  ├── 👤 Table: users                                                           │
│  │   ├── id (Primary Key)                                                     │
│  │   ├── name, email, mobile                                                  │
│  │   ├── password_hash                                                        │
│  │   ├── registration_status                                                  │
│  │   └── profile_completed                                                    │
│  │                                                                             │
│  └── 📍 Table: user_addresses                                                  │
│      ├── id (Primary Key)                                                     │
│      ├── user_id (Foreign Key → users.id)                                     │
│      ├── label VARCHAR(50)                                                    │
│      ├── address_line TEXT                                                    │
│      ├── city VARCHAR(100)                                                    │
│      ├── state VARCHAR(100)                                                   │
│      ├── zip_code VARCHAR(20)                                                 │
│      ├── country VARCHAR(100)                                                 │
│      ├── latitude DECIMAL(10,8)                                               │
│      ├── longitude DECIMAL(11,8)                                              │
│      ├── is_default BOOLEAN                                                   │
│      ├── created_at TIMESTAMP                                                 │
│      └── updated_at TIMESTAMP                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Data Flow Sequence

### 1. Page Load & Authentication
```
User → addresses.html → addresses-api.js → Check Auth Token
                                        ↓
                                   Token Valid?
                                        ↓
                              YES: Load Addresses
                              NO: Redirect to Login
```

### 2. Loading Addresses
```
Frontend → GET /api/v1/users/addresses
         ↓
API Gateway → user-service → AuthMiddleware → GetAddresses()
                                           ↓
                                    SELECT * FROM user_addresses 
                                    WHERE user_id = ? 
                                    ORDER BY is_default DESC
                                           ↓
                                    Return JSON Array
```

### 3. Creating New Address
```
User Fills Form → addresses-api.js → POST /api/v1/users/addresses
                                   ↓
                            Validation & Auth Check
                                   ↓
                            CreateAddress() Handler
                                   ↓
                    If is_default = true: UPDATE other addresses
                                   ↓
                    INSERT INTO user_addresses (...)
                                   ↓
                            Log Activity & Return Success
```

### 4. Map Integration & GPS Coordinates
```
User Clicks "Select on Map" → Open Leaflet Map Modal
                            ↓
                    User Clicks Location on Map
                            ↓
                    Reverse Geocoding (OpenStreetMap)
                            ↓
                    Auto-fill Address Fields
                            ↓
                    Store Latitude/Longitude in Form
                            ↓
                    Save with GPS Coordinates
```

## Key Features & Data Storage

### 🔐 Authentication & Security
- **JWT Token**: Stored in localStorage/sessionStorage
- **Device ID**: Generated and stored for tracking
- **Session Management**: Redis-based session cleanup
- **Middleware Protection**: All address endpoints require authentication

### 📍 Address Data Structure
```json
{
  "id": 123,
  "user_id": 456,
  "label": "Home",
  "address_line": "123 Main Street, Apt 4B",
  "city": "New York",
  "state": "NY",
  "zip_code": "10001",
  "country": "USA",
  "latitude": 40.7128,
  "longitude": -74.0060,
  "is_default": true,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### 🗺️ Map Integration
- **Map Provider**: OpenStreetMap via Leaflet.js
- **Geocoding**: Nominatim reverse geocoding service
- **GPS Support**: Browser geolocation API
- **Interactive**: Click to place marker, drag to adjust

### 💾 Data Persistence
- **Primary Storage**: PostgreSQL database
- **Indexes**: Optimized queries on user_id and is_default
- **Constraints**: Foreign key relationship with users table
- **Migration**: Automatic migration from localStorage to database

### 🔄 State Management
- **Default Address**: Only one per user (enforced at database level)
- **Activity Logging**: All CRUD operations logged for audit
- **Error Handling**: Comprehensive error responses and user feedback
- **Offline Support**: LocalStorage fallback during migration

## API Endpoints Summary

| Method | Endpoint | Purpose | Auth Required |
|--------|----------|---------|---------------|
| GET | `/api/v1/users/addresses` | List all user addresses | ✅ |
| POST | `/api/v1/users/addresses` | Create new address | ✅ |
| PUT | `/api/v1/users/addresses/:id` | Update existing address | ✅ |
| DELETE | `/api/v1/users/addresses/:id` | Delete address | ✅ |
| PUT | `/api/v1/users/addresses/:id/default` | Set as default address | ✅ |

## Database Relationships

```
users (1) ←→ (many) user_addresses
  ↓                    ↓
user.id ←────────→ user_addresses.user_id
                       ↓
              CASCADE DELETE enabled
```

This architecture ensures secure, scalable address management with GPS integration, real-time validation, and comprehensive audit logging.