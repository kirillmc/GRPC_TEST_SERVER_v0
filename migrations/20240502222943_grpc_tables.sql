-- +goose Up
create table users
(
    id            serial primary key,
    name          varchar(50)      not null,
    surname       varchar(50)      not null,
    email         varchar(255)     not null,
    avatar        varchar(255)     not null,
    login         varchar(50)      not null,
    password varchar(255)     not null,
    role          integer          not null,
    weight        double precision not null,
    height        double precision not null,
    locked        boolean          not null
);

-- +goose Down
drop table users;
