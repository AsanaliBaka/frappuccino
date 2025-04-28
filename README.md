<p align="center">
  <a href="" rel="noopener">
  <img width=200px height=200px src="logo.png" alt="Project logo" style="border-radius:0%;"></a>
  
</p>
<h3 align="center">FRAPPUCINO</h3>

<div align="center">

[![Status](https://img.shields.io/badge/status-completed-success.svg)]()
<br>
![Started](https://img.shields.io/date/1742899680?label=Started)
![Last commit](https://img.shields.io/date/1743719680?label=Last%20commit)

## Presentation
<div align="center">
  <img src="assets/introgif.gif" alt="Demo GIF" />
</div>
---

<div align="left">

# 📦 Frappuccino

Frappuccino — REST API for managing orders, menu, inventory, and analytics, inspired by Amazon S3.

## 🚀 Quick Start

```bash
docker-compose up --build
```

The server starts at `http://localhost:8080`

## 📮 POST Requests

### 🛒 /orders
Create a new order.
```json
{
  "customer_id": "1",
  "status": "open",
  "positions": [
    {
      "item_id": "4",
      "count": 2,
      "unit_price": 14.75,
      "adjustments": { "extra_cheese": true }
    }
  ],
  "amount": 29.50,
  "note": "Extra hot"
}
```

### 🧾 /orders/batch
Submit a batch of orders.
```json
{
  "purchases": [ /* array of order objects */ ]
}
```

### ✅ /orders/{id}/close
Close an existing order by ID.

### 🍽 /menu
Create a new menu item.
```json
{
  "title": "Borscht",
  "details": "Beet soup with sour cream",
  "unit_price": 7.99,
  "size_label": "medium",
  "group": "Soup",
  "labels": ["beet", "sour_cream"],
  "extras": {"vegetarian": true},
  "components": [
    { "component_id": "1", "required_qty": 0.3 }
  ]
}
```

### 🧂 /inventory
Add a new inventory item.
```json
{
  "title": "Beetroot",
  "stock": 50,
  "measure": "kg",
  "unit_cost": 2.5
}
```

---

## 📊 Analytics

- `GET /orders/number` — total sold items within a period
- `GET /reports/total-sales` — total revenue
- `GET /reports/popular-items` — most popular dishes

## 🗃 Stack
- Go
- PostgreSQL
- Docker

## ✍🏻 Authors <a name = "authors"></a>

- [![Status](https://img.shields.io/badge/alem-azhaxyly-success?logo=github)](https://platform.alem.school/git/azhaxyly) <a href="https://t.me/hmlssdeus" target="_blank"><img src="https://img.shields.io/badge/telegram-@hmlssdeus-blue?logo=Telegram" alt="Status" /></a>
- [![Status](https://img.shields.io/badge/alem-abaltash-success?logo=github)](https://platform.alem.school/git/abaltash)


## 🎉 Acknowledgements <a name = "acknowledgement"></a>

