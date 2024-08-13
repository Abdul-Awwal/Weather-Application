import React from 'react';

function WeatherDisplay({ weather }) {
    //Display weather information to the user
    return (
        <div>
            <h2>Weather Information</h2>
            <p>Temperature: {weather.temperature}</p>
            <p>Humidity: {weather.humidity}</p>
            <p>Wind Speed: {weather.windSpeed}</p>
            <p>Conditions: {weather.conditions}</p>
        </div>
    );
}

export default WeatherDisplay;
