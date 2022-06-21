-- create database pet_api;

drop table if exists kat;

set timezone = 'Europe/Amsterdam';

create table if not exists kat (
    id serial primary key,
    name text,
    created_at timestamp,
    updated_at timestamp
);

insert into kat (name, created_at, updated_at) values ('Melinoe', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
insert into kat (name, created_at, updated_at) values ('Salem', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
insert into kat (name, created_at, updated_at) values ('Kicks', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
insert into kat (name, created_at, updated_at) values ('Pablo', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
insert into kat (name, created_at, updated_at) values ('Chili', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

select * from kat;