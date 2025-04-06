-- +goose Up
-- +goose StatementBegin
create or replace function goose_manage_updated_at(_tbl regclass) returns void as $$
begin
    execute format('create trigger set_updated_at before update on %s
                    for each row execute procedure goose_set_updated_at()', _tbl);
end;
$$ language plpgsql;

create or replace function goose_set_updated_at() returns trigger as $$
begin
    if (
        new is distinct from old and
        new.updated_at is not distinct from old.updated_at
    ) then
        new.updated_at := current_timestamp;
    end if;
    return new;
end;
$$ language plpgsql;

create extension if not exists "uuid-ossp";
create collation case_insensitive (provider = icu, locale = 'und-u-ks-level2');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop function if exists goose_manage_updated_at(_tbl regclass);
drop function if exists goose_set_updated_at();
drop collation if exists case_insensitive;
drop extension if exists "uuid-ossp";
-- +goose StatementEnd
