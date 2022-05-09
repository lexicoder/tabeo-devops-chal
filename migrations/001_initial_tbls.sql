CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE events(
    id UUID PRIMARY KEY,
    name VARCHAR(100),
    ts TIMESTAMP WITH TIME ZONE
);

CREATE TABLE destinations(
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE users(
    id UUID PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    gender CHAR(1) NOT NULL,
    birthday DATE NOT NULL
);

CREATE OR REPLACE FUNCTION launch_in_same_week(pad CHAR(24), dest UUID, d DATE)
RETURNS BOOLEAN 
language plpgsql
AS
$$
BEGIN
	PERFORM A.id FROM flights A WHERE A.launchpad_id = pad AND A.destination_id = dest AND date_part('week', A.launch_date) = date_part('week', d) LIMIT 1;
	IF FOUND
		THEN RETURN false;
		ELSE RETURN true;
	END IF;
END;
$$;

CREATE TABLE flights(
    id UUID PRIMARY KEY,
    launchpad_id CHAR(24) NOT NULL,
    destination_id UUID NOT NULL,
    launch_date DATE NOT NULL,
    UNIQUE(launchpad_id, launch_date),
    CONSTRAINT fk_destination FOREIGN KEY(destination_id) REFERENCES destinations(id) ON DELETE CASCADE,
    CONSTRAINT check_unique_launchpad_dest_in_week CHECK(launch_in_same_week(launchpad_id, destination_id, launch_date))
);

CREATE TABLE bookings(
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    flight_id UUID NOT NULL,
    status VARCHAR(10) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(user_id, flight_id),
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_flight FOREIGN KEY(flight_id) REFERENCES flights(id) ON DELETE CASCADE
);

CREATE INDEX idx_bookings_pagination ON bookings (created_at, id);

INSERT INTO destinations VALUES 
('05c7f2ca-aa9a-4ea8-a6d5-4cb691468830', 'Mars'), 
('88aed240-f3f5-4a21-8968-718e08f27c68', 'Moon'),
('c1f4cbcc-5df9-41f7-9486-3cb2103d1262', 'Pluto'),
('1b3bab7f-9efa-4727-a308-f98775d807df', 'Asteroid Belt'),
('d6e75ca7-1737-4cb7-a648-9375a5b28055', 'Europa'),
('e0ea6dc2-0c71-41e1-9bcb-e4f843a2736f', 'Titan'),
('03f719a1-aa1a-4e85-9e3d-8b455f10a9f4', 'Ganymede');

---- create above / drop below ----

DROP TABLE bookings;
DROP TABLE flights;
DROP FUNCTION launch_in_same_week;
DROP TABLE users;
DROP TABLE destinations;
