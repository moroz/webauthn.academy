-- +goose Up
-- +goose StatementBegin
create extension if not exists citext;

create table users (
  id bigint primary key generated by default as identity,
  email citext not null unique,
  display_name varchar(80) not null,
  password_hash varchar(100) not null,
  inserted_at timestamp(0) not null default (now() at time zone 'utc'),
  updated_at timestamp(0) not null default (now() at time zone 'utc')
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table users;
-- +goose StatementEnd
