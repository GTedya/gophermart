START TRANSACTION;

CREATE TABLE IF NOT EXISTS order_accruals
(
    id          SERIAL PRIMARY KEY,
    order_id    VARCHAR(50),
    user_id     INT,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    status      VARCHAR(50) DEFAULT 'NEW',
    accrual     DECIMAL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

COMMIT;
