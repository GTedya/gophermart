START TRANSACTION;

CREATE TYPE order_status AS ENUM ('NEW','PROCESSED' ,'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE IF NOT EXISTS order_accruals
(
    id          SERIAL PRIMARY KEY,
    order_id    BIGINT,
    user_id     INT,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    status      order_status DEFAULT 'NEW',
    accrual     DECIMAL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

COMMIT;
