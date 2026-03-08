CREATE SCHEMA IF NOT EXISTS meteo;

CREATE TABLE IF NOT EXISTS meteo.weather (
    id          BIGSERIAL PRIMARY KEY,
    time        TIMESTAMPTZ NOT NULL,
    city        TEXT NOT NULL,
    temperature REAL NOT NULL,
    pressure    REAL NOT NULL,
    humidity    REAL NOT NULL,
    wind_speed  REAL NOT NULL,
    uv          REAL NOT NULL,
    description TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_weather_city_time ON meteo.weather (city, time DESC);
