/*
CREATE USER lk WITH password 'liokor';
CREATE DATABASE liokoredu OWNER lk;
GRANT ALL privileges ON DATABASE liokoredu TO lk;
*/

CREATE TABLE users(    id       bigserial primary key,    name    varchar(60) not null);
 insert into users values (1, 'admin');

CREATE TABLE tasks
(
    id          bigserial primary key,
    title text not null, 
    description  text not null,
    hints       text,
    input text not null,
    output text not null, 
    test_amount int not null default 0,
    tests     text not null,
    creator bigint references users (id) on delete cascade,
    is_private boolean not null default false,
    code varchar(10) default null,
    date timestamp not null
);

/*
CREATE TABLE solutions
(
    id          bigserial primary key,
    receiveDateTime  timestamp not null,
    checkDateTime  timestamp not null,
    status       integer,
    test_examples      text not null,
    creator bigint references users (id) on delete cascade,
    is_private boolean not null ,
    code varchar(10) default null,
    date timestamp not null, 
);
insert into tasks values (1, 'Sum of two numbers',         'Given two numbers. Calculate their sum.',         'It is a very easy task',         '-2^31 <= a <= 2^32, -2^31 <= b <= 2^32',         'c = a + b', 10, 
        '[["1 2", "3"],["-2 2", "0"]]', 1, false,null, '2021-05-31 19:00:00');sudo su postgres
psql 

psql -c "CREATE USER lk WITH superuser login password 'liokor';"
    psql -c "CREATE USER postgre WITH superuser login password 'postgre';" && \
    psql -c "ALTER ROLE postgre WITH PASSWORD 'postgre';" && \
    createdb -O android moviedb && \
    psql -d proxy < ./database.sql && \
    service postgresql stop
    */