-- +goose up 
alter table users add column if not exists password text not null;


-- +goose down
alter table users drop column password;
