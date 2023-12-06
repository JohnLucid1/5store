CREATE TABLE Links (
    data_id serial primary key, 
    data text not null unique, 
    created_url varchar(50) not null unique,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
