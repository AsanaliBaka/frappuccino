CREATE TYPE order_status AS ENUM ('canceled', 'completed', 'open');
CREATE TYPE size_type AS ENUM ('small', 'medium', 'large', 'extra_large');
CREATE TYPE unit_type AS ENUM ('kg', 'l', 'pcs');
CREATE TYPE transaction_type AS ENUM ('addition', 'consumption');

CREATE TABLE Customers (
    Customer_ID SERIAL PRIMARY KEY,
    Name VARCHAR(255) NOT NULL,
    Email VARCHAR(255) UNIQUE NOT NULL,
    Phone VARCHAR(20) NOT NULL,
    Preference JSONB DEFAULT '{}'::JSONB
);

CREATE TABLE Orders (
    Order_ID SERIAL PRIMARY KEY,
    Status order_status NOT NULL,
    Total_Amount DECIMAL(10, 2) NOT NULL,
    Created_At TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    Updated_At TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    Customer_ID INTEGER NOT NULL,
    FOREIGN KEY (Customer_ID) REFERENCES Customers(Customer_ID) ON DELETE CASCADE
);

CREATE TABLE Order_Status_History (
    History_ID SERIAL PRIMARY KEY,
    Order_ID INTEGER NOT NULL,
    Status order_status NOT NULL,
    Changed_At TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (Order_ID) REFERENCES Orders(Order_ID) ON DELETE CASCADE
);

CREATE TABLE Menu_Items (
    Menu_Item_ID SERIAL PRIMARY KEY,
    Name VARCHAR(100) NOT NULL UNIQUE,
    Description TEXT,
    Price DECIMAL(10, 2) NOT NULL,
    Size size_type,
    Category VARCHAR(50) NOT NULL,
    Tags TEXT[],
    Metadata JSONB DEFAULT '{}'::JSONB
);

CREATE TABLE Order_Items (
    Order_Item_ID SERIAL PRIMARY KEY,
    Order_ID INTEGER NOT NULL,
    Menu_Item_ID INTEGER,
    Quantity DECIMAL(10,3) NOT NULL,
    Price DECIMAL(10, 2) NOT NULL,
    Customization JSONB DEFAULT '{}'::JSONB,
    Created_At TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (Order_ID) REFERENCES Orders(Order_ID) ON DELETE CASCADE,
    FOREIGN KEY (Menu_Item_ID) REFERENCES Menu_Items(Menu_Item_ID) ON DELETE SET NULL
);

CREATE TABLE Price_History (
    Price_ID SERIAL PRIMARY KEY,
    Menu_Item_ID INTEGER NOT NULL,
    Old_Price DECIMAL(10, 2) NOT NULL,
    New_Price DECIMAL(10, 2) NOT NULL,
    Changed_At TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (Menu_Item_ID) REFERENCES Menu_Items(Menu_Item_ID) ON DELETE CASCADE
);

CREATE TABLE Inventory (
    Inventory_ID SERIAL PRIMARY KEY,
    Name VARCHAR(255) NOT NULL,
    Quantity DECIMAL(12,4) NOT NULL,
    Unit unit_type NOT NULL,
    Price NUMERIC(10, 2) NOT NULL
);

CREATE TABLE Menu_Item_Ingredients (
    Menu_Item_Ingredients_ID SERIAL PRIMARY KEY,
    Menu_Item_ID INTEGER NOT NULL,
    Inventory_ID INTEGER NOT NULL,
    Quantity DECIMAL(10,3) NOT NULL,
    FOREIGN KEY (Menu_Item_ID) REFERENCES Menu_Items(Menu_Item_ID) ON DELETE CASCADE,
    FOREIGN KEY (Inventory_ID) REFERENCES Inventory(Inventory_ID) ON DELETE CASCADE
);

CREATE TABLE Inventory_Transactions (
    Transaction_ID SERIAL PRIMARY KEY,
    Inventory_ID INTEGER NOT NULL,
    Change_Amount DECIMAL(12, 4) NOT NULL,
    Transaction_Type transaction_type NOT NULL,
    Occurred_At TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (Inventory_ID) REFERENCES Inventory(Inventory_ID) ON DELETE CASCADE
);

CREATE TABLE Inventory_Reservations (
    Reservation_ID SERIAL PRIMARY KEY,
    Order_ID INTEGER NOT NULL,
    Inventory_ID INTEGER NOT NULL,
    Reserved_Quantity DECIMAL(12, 4) NOT NULL,
    Created_At TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (Order_ID) REFERENCES Orders(Order_ID) ON DELETE CASCADE,
    FOREIGN KEY (Inventory_ID) REFERENCES Inventory(Inventory_ID) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_order_items_unique ON Order_Items (Order_ID, Menu_Item_ID);

CREATE INDEX idx_orders_customer_id ON Orders(Customer_ID);

CREATE INDEX idx_inventory_transactions_inventory_id ON Inventory_Transactions(Inventory_ID);

CREATE INDEX idx_order_status_history_order_id ON Order_Status_History(Order_ID);

CREATE INDEX idx_orders_status ON Orders(Status);

CREATE INDEX idx_menu_item_ingredients_composite ON Menu_Item_Ingredients(Menu_Item_ID, Inventory_ID);

CREATE OR REPLACE FUNCTION log_inventory_transaction()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO Inventory_Transactions (Inventory_ID, Change_Amount, Transaction_Type, Occurred_At)
        VALUES (NEW.Inventory_ID, NEW.Quantity, 'addition'::transaction_type, NOW());

    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO Inventory_Transactions (Inventory_ID, Change_Amount, Transaction_Type, Occurred_At)
        VALUES (
            NEW.Inventory_ID,
            NEW.Quantity - OLD.Quantity,
            CASE 
                WHEN NEW.Quantity > OLD.Quantity THEN 'addition'::transaction_type
                ELSE 'consumption'::transaction_type 
            END,
            NOW()
        );
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE TRIGGER inventory_transaction_trigger
AFTER INSERT OR UPDATE ON Inventory
FOR EACH ROW
EXECUTE FUNCTION log_inventory_transaction();


-- Customers
INSERT INTO Customers (Customer_ID, Name, Email, Phone)
VALUES
(1, 'Alice Johnson', 'alice@example.com', '123-456-7890'),
(2, 'Bob Smith', 'bob@example.com', '234-567-8901'),
(3, 'Carol Williams', 'carol@example.com', '345-678-9012'),
(4, 'David Brown', 'david@example.com', '456-789-0123'),
(5, 'Eva Davis', 'eva@example.com', '567-890-1234');

-- Menu Items
INSERT INTO Menu_Items (Name, Description, Price, Size, Category, Tags, Metadata) VALUES
('Caesar Salad', 'Fresh romaine lettuce with grilled chicken and parmesan cheese', 8.99, 'medium', 'Appetizer', ARRAY['salad', 'chicken'], '{"spicy": false}'),
('Bruschetta', 'Grilled bread rubbed with garlic and topped with tomatoes, basil, and olive oil', 6.50, 'small', 'Appetizer', ARRAY['bread', 'tomato'], '{"vegan": true}'),
('Pumpkin Soup', 'Creamy pumpkin soup with a hint of nutmeg', 7.25, 'small', 'Appetizer', ARRAY['soup', 'pumpkin'], '{"gluten_free": true}'),
('Beef Steak', 'Grilled beef steak served with seasonal vegetables', 18.50, 'large', 'Main Course', ARRAY['steak', 'beef'], '{"doneness": "medium rare"}'),
('Pasta Carbonara', 'Classic Italian pasta with egg, cheese, and pancetta', 14.75, 'medium', 'Main Course', ARRAY['pasta', 'italian'], '{"contains_pork": true}'),
('Grilled Salmon', 'Fresh salmon grilled to perfection with lemon butter sauce', 16.00, 'large', 'Main Course', ARRAY['fish', 'salmon'], '{"omega3": true}'),
('Tiramisu', 'Traditional Italian dessert with layers of coffee-soaked ladyfingers and mascarpone', 6.00, 'medium', 'Dessert', ARRAY['dessert', 'coffee'], '{"contains_alcohol": false}'),
('Cheesecake', 'New York style cheesecake with a graham cracker crust', 6.25, 'medium', 'Dessert', ARRAY['dessert', 'cheese'], '{"sweetness": "high"}'),
('Fruit Salad', 'A mix of seasonal fruits served fresh', 5.50, 'small', 'Dessert', ARRAY['dessert', 'fruit'], '{"vegan": true}'),
('Chocolate Lava Cake', 'Warm chocolate cake with a molten chocolate center', 7.50, 'small', 'Dessert', ARRAY['dessert', 'chocolate'], '{"temperature": "warm"}');

-- Price History
INSERT INTO Price_History (Menu_Item_ID, Old_Price, New_Price, Changed_At) VALUES
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Caesar Salad'), 7.50, 8.99, '2025-01-10 10:00:00'),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Beef Steak'), 16.50, 18.50, '2025-01-20 11:00:00'),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Pasta Carbonara'), 13.50, 14.75, '2025-01-15 09:30:00'),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Tiramisu'), 5.00, 6.00, '2025-01-05 08:45:00');

-- Inventory
INSERT INTO Inventory (Name, Quantity, Unit, Price) VALUES
('Romaine Lettuce', 50.0000, 'kg', 1.50),
('Chicken Breast', 30.0000, 'kg', 5.00),
('Parmesan Cheese', 20.0000, 'kg', 8.00),
('Croutons', 15.0000, 'kg', 2.50),
('Tomatoes', 40.0000, 'kg', 1.20),
('Garlic', 10.0000, 'kg', 3.00),
('Pumpkin', 25.0000, 'kg', 1.80),
('Nutmeg', 5.0000, 'kg', 15.00),
('Beef', 40.0000, 'kg', 12.00),
('Mixed Vegetables', 60.0000, 'kg', 2.00),
('Pasta', 35.0000, 'kg', 3.50),
('Eggs', 200.0000, 'pcs', 0.30),
('Pancetta', 15.0000, 'kg', 7.00),
('Salmon', 20.0000, 'kg', 10.00),
('Ladyfingers', 10.0000, 'kg', 4.00),
('Mascarpone', 8.0000, 'kg', 6.00),
('Graham Crackers', 12.0000, 'kg', 3.50),
('Chocolate', 25.0000, 'kg', 5.00),
('Fruits', 50.0000, 'kg', 2.50),
('Lemon', 100.0000, 'pcs', 0.25);

-- Menu Item Ingredients
INSERT INTO Menu_Item_Ingredients (Menu_Item_ID, Inventory_ID, Quantity) VALUES
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Caesar Salad'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Romaine Lettuce'), 0.200),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Caesar Salad'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Parmesan Cheese'), 0.050),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Bruschetta'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Tomatoes'), 0.150),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Bruschetta'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Garlic'), 0.020),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Pumpkin Soup'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Pumpkin'), 0.300),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Pumpkin Soup'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Nutmeg'), 0.005),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Beef Steak'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Beef'), 0.400),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Beef Steak'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Mixed Vegetables'), 0.250),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Pasta Carbonara'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Pasta'), 0.300),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Pasta Carbonara'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Eggs'), 0.050),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Grilled Salmon'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Salmon'), 0.350),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Grilled Salmon'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Lemon'), 0.100),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Tiramisu'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Ladyfingers'), 0.200),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Tiramisu'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Mascarpone'), 0.150),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Cheesecake'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Graham Crackers'), 0.250),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Cheesecake'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Parmesan Cheese'), 0.030),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Fruit Salad'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Fruits'), 0.300),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Fruit Salad'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Tomatoes'), 0.150),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Chocolate Lava Cake'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Chocolate'), 0.200),
((SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Chocolate Lava Cake'), (SELECT Inventory_ID FROM Inventory WHERE Name = 'Eggs'), 0.040);

-- Inventory Transactions
INSERT INTO Inventory_Transactions (Inventory_ID, Change_Amount, Transaction_Type, Occurred_At) VALUES
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Romaine Lettuce'), 20.0000, 'addition', '2025-02-01 07:00:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Chicken Breast'), 15.0000, 'addition', '2025-02-01 07:05:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Parmesan Cheese'), 10.0000, 'addition', '2025-02-01 07:10:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Croutons'), 8.0000, 'addition', '2025-02-01 07:15:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Tomatoes'), 25.0000, 'addition', '2025-02-01 07:20:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Garlic'), 5.0000, 'addition', '2025-02-01 07:25:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Pumpkin'), 12.0000, 'addition', '2025-02-01 07:30:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Nutmeg'), 3.0000, 'addition', '2025-02-01 07:35:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Beef'), 20.0000, 'addition', '2025-02-01 07:40:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Mixed Vegetables'), 15.0000, 'addition', '2025-02-01 07:45:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Pasta'), 10.0000, 'addition', '2025-02-01 07:50:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Eggs'), 50.0000, 'addition', '2025-02-01 07:55:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Pancetta'), 8.0000, 'addition', '2025-02-01 08:00:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Salmon'), 12.0000, 'addition', '2025-02-01 08:05:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Ladyfingers'), 5.0000, 'addition', '2025-02-01 08:10:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Mascarpone'), 7.0000, 'addition', '2025-02-01 08:15:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Graham Crackers'), 10.0000, 'addition', '2025-02-01 08:20:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Chocolate'), 4.0000, 'addition', '2025-02-01 08:25:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Fruits'), 15.0000, 'addition', '2025-02-01 08:30:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Lemon'), 30.0000, 'addition', '2025-02-01 08:35:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Chicken Breast'), -5.0000, 'consumption', '2025-02-01 09:00:00'),
((SELECT Inventory_ID FROM Inventory WHERE Name = 'Croutons'), -2.0000, 'consumption', '2025-02-01 09:05:00');

-- Orders
INSERT INTO Orders (Status, Total_Amount, Created_At, Updated_At, Customer_ID) VALUES
('open',  25.50, '2025-02-01 12:15:00', '2025-02-01 12:15:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Alice Johnson')),
('completed', 40.00, '2025-02-01 12:30:00', '2025-02-01 12:45:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Bob Smith')),
('canceled',  15.75, '2025-02-01 12:45:00', '2025-02-01 12:50:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Carol Williams')),
('completed', 30.00, '2025-02-01 13:00:00', '2025-02-01 13:10:00', (SELECT Customer_ID FROM Customers WHERE Name = 'David Brown')),
('canceled', 12.50, '2025-02-01 13:15:00', '2025-02-01 13:20:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Eva Davis')),
('completed', 22.00, '2025-02-02 11:00:00', '2025-02-02 11:30:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Alice Johnson')),
('open', 18.75, '2025-02-02 11:30:00', '2025-02-02 11:30:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Bob Smith')),
('completed', 27.50, '2025-02-02 11:45:00', '2025-02-02 12:00:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Carol Williams')),
('open', 35.00, '2025-02-02 12:00:00', '2025-02-02 12:00:00', (SELECT Customer_ID FROM Customers WHERE Name = 'David Brown')),
('canceled', 16.25, '2025-02-02 12:15:00', '2025-02-02 12:20:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Eva Davis')),
('open', 20.00, '2025-02-03 13:20:00', '2025-02-03 13:20:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Alice Johnson')),
('completed', 45.50, '2025-02-03 13:35:00', '2025-02-03 13:50:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Bob Smith')),
('canceled', 10.00, '2025-02-03 13:50:00', '2025-02-03 13:55:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Carol Williams')),
('completed', 33.75, '2025-02-03 14:05:00', '2025-02-03 14:15:00', (SELECT Customer_ID FROM Customers WHERE Name = 'David Brown')),
('canceled', 14.50, '2025-02-03 14:20:00', '2025-02-03 14:25:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Eva Davis')),
('open', 28.00, '2025-02-04 12:10:00', '2025-02-04 12:10:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Alice Johnson')),
('completed', 38.25, '2025-02-04 12:25:00', '2025-02-04 12:40:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Bob Smith')),
('canceled', 19.00, '2025-02-04 12:40:00', '2025-02-04 12:45:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Carol Williams')),
('completed', 31.50, '2025-02-04 12:55:00', '2025-02-04 13:05:00', (SELECT Customer_ID FROM Customers WHERE Name = 'David Brown')),
('canceled', 17.25, '2025-02-04 13:10:00', '2025-02-04 13:15:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Eva Davis')),
('completed', 26.00, '2025-02-05 11:05:00', '2025-02-05 11:15:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Alice Johnson')),
('open', 19.75, '2025-02-05 11:20:00', '2025-02-05 11:20:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Bob Smith')),
('canceled', 11.00, '2025-02-05 11:35:00', '2025-02-05 11:40:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Carol Williams')),
('completed', 34.50, '2025-02-05 11:50:00', '2025-02-05 12:00:00', (SELECT Customer_ID FROM Customers WHERE Name = 'David Brown')),
('canceled', 12.75, '2025-02-05 12:05:00', '2025-02-05 12:10:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Eva Davis')),
('open', 29.00, '2025-02-06 13:00:00', '2025-02-06 13:00:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Alice Johnson')),
('completed', 37.50, '2025-02-06 13:15:00', '2025-02-06 13:30:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Bob Smith')),
('canceled', 14.00, '2025-02-06 13:30:00', '2025-02-06 13:35:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Carol Williams')),
('open', 32.25, '2025-02-06 13:45:00', '2025-02-06 13:45:00', (SELECT Customer_ID FROM Customers WHERE Name = 'David Brown')),
('completed', 20.50, '2025-02-06 14:00:00', '2025-02-06 14:00:00', (SELECT Customer_ID FROM Customers WHERE Name = 'Eva Davis'));

--Order Status History
-- Order 1 (open)
INSERT INTO Order_Status_History (Order_ID, Status, Changed_At)
VALUES ((SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-01 12:15:00'), 'open', '2025-02-01 12:15:00');

-- Order 2 (open → completed)
INSERT INTO Order_Status_History (Order_ID, Status, Changed_At)
VALUES 
((SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-01 12:30:00'), 'open', '2025-02-01 12:30:00'),
((SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-01 12:30:00'), 'completed', '2025-02-01 12:45:00');

-- Order 3 (open → canceled)
INSERT INTO Order_Status_History (Order_ID, Status, Changed_At)
VALUES 
((SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-01 12:45:00'), 'open', '2025-02-01 12:45:00'),
((SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-01 12:45:00'), 'canceled', '2025-02-01 12:50:00');

-- Order 4 (open → completed)
INSERT INTO Order_Status_History (Order_ID, Status, Changed_At)
VALUES 
((SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-01 13:00:00'), 'open', '2025-02-01 13:00:00'),
((SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-01 13:00:00'), 'completed', '2025-02-01 13:10:00');

-- Order 5 (open → canceled)
INSERT INTO Order_Status_History (Order_ID, Status, Changed_At)
VALUES 
((SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-01 13:15:00'), 'open', '2025-02-01 13:15:00'),
((SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-01 13:15:00'), 'canceled', '2025-02-01 13:20:00');

-- Order 6 (open → completed)
INSERT INTO Order_Status_History (Order_ID, Status, Changed_At)
VALUES 
((SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-02 11:00:00'), 'open', '2025-02-02 11:00:00'),
((SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-02 11:00:00'), 'completed', '2025-02-02 11:30:00');

-- Order 7 (open)
INSERT INTO Order_Status_History (Order_ID, Status, Changed_At)
VALUES ((SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-02 11:30:00'), 'open', '2025-02-02 11:30:00');

-- Order 8 (open → completed)
INSERT INTO Order_Status_History (Order_ID, Status, Changed_At)
VALUES 
((SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-02 11:45:00'), 'open', '2025-02-02 11:45:00'),
((SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-02 11:45:00'), 'completed', '2025-02-02 12:00:00');

-- Order 9 (open)
INSERT INTO Order_Status_History (Order_ID, Status, Changed_At)
VALUES ((SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-02 12:00:00'), 'open', '2025-02-02 12:00:00');

-- Order 10 (open → canceled)
INSERT INTO Order_Status_History (Order_ID, Status, Changed_At)
VALUES 
((SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-02 12:15:00'), 'open', '2025-02-02 12:15:00'),
((SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-02 12:15:00'), 'canceled', '2025-02-02 12:20:00');

-- Order_Items
-- Order 1: Caesar Salad
INSERT INTO Order_Items (Order_ID, Menu_Item_ID, Quantity, Price, Customization, Created_At)
VALUES (
    (SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-01 12:15:00'),
    (SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Caesar Salad'),
    1.000, 8.99, '{}', '2025-02-01 12:16:00'
);

-- Order 2: Beef Steak
INSERT INTO Order_Items (Order_ID, Menu_Item_ID, Quantity, Price, Customization, Created_At)
VALUES (
    (SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-01 12:30:00'),
    (SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Beef Steak'),
    1.000, 18.50, '{}', '2025-02-01 12:31:00'
);

-- Order 3: Bruschetta
INSERT INTO Order_Items (Order_ID, Menu_Item_ID, Quantity, Price, Customization, Created_At)
VALUES (
    (SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-01 12:45:00'),
    (SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Bruschetta'),
    1.000, 6.50, '{}', '2025-02-01 12:46:00'
);

-- Order 4: Pasta Carbonara
INSERT INTO Order_Items (Order_ID, Menu_Item_ID, Quantity, Price, Customization, Created_At)
VALUES (
    (SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-01 13:00:00'),
    (SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Pasta Carbonara'),
    1.000, 14.75, '{}', '2025-02-01 13:01:00'
);

-- Order 5: Pumpkin Soup
INSERT INTO Order_Items (Order_ID, Menu_Item_ID, Quantity, Price, Customization, Created_At)
VALUES (
    (SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-01 13:15:00'),
    (SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Pumpkin Soup'),
    1.000, 7.25, '{}', '2025-02-01 13:16:00'
);

-- Order 6: Caesar Salad
INSERT INTO Order_Items (Order_ID, Menu_Item_ID, Quantity, Price, Customization, Created_At)
VALUES (
    (SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-02 11:00:00'),
    (SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Caesar Salad'),
    1.000, 8.99, '{}', '2025-02-02 11:01:00'
);

-- Order 7: Grilled Salmon
INSERT INTO Order_Items (Order_ID, Menu_Item_ID, Quantity, Price, Customization, Created_At)
VALUES (
    (SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-02 11:30:00'),
    (SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Grilled Salmon'),
    1.000, 16.00, '{}', '2025-02-02 11:31:00'
);

-- Order 8: Tiramisu
INSERT INTO Order_Items (Order_ID, Menu_Item_ID, Quantity, Price, Customization, Created_At)
VALUES (
    (SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-02 11:45:00'),
    (SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Tiramisu'),
    1.000, 6.00, '{}', '2025-02-02 11:46:00'
);

-- Order 9: Cheesecake
INSERT INTO Order_Items (Order_ID, Menu_Item_ID, Quantity, Price, Customization, Created_At)
VALUES (
    (SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-02 12:00:00'),
    (SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Cheesecake'),
    1.000, 6.25, '{}', '2025-02-02 12:01:00'
);

-- Order 10: Fruit Salad (с кастомизацией)
INSERT INTO Order_Items (Order_ID, Menu_Item_ID, Quantity, Price, Customization, Created_At)
VALUES (
    (SELECT Order_ID FROM Orders WHERE Created_At = '2025-02-02 12:15:00'),
    (SELECT Menu_Item_ID FROM Menu_Items WHERE Name = 'Fruit Salad'),
    1.000, 5.50, '{"extra_toppings": ["honey", "nuts"], "size": "large"}', '2025-02-02 12:16:00'
);