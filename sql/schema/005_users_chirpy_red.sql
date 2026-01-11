-- +goose up
alter table users add column if not exists is_chirpy_red bool not null default false;

-- +goose down
alter table users drop column is_chirpy_red;
