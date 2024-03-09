START TRANSACTION;

CREATE TABLE IF NOT EXISTS write_off_history
(
    id        SERIAL PRIMARY KEY ,
    order_id  BIGINT,
    user_id   INT,
    withdrawn DECIMAL,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    FOREIGN KEY (user_id) REFERENCES users (id)
);

COMMIT;
