import React, { useState } from 'react';

function WeatherForm({ setWeather }) {
    // State variables to manage user input and error messages
    const [location, setLocation] = useState('');
    const [latitude, setLatitude] = useState('');
    const [longitude, setLongitude] = useState('');
    const [useCoordinates, setUseCoordinates] = useState(false);
    const [error, setError] = useState('');

    /* handleSubmit takes the user input and uses that to make a request to the backend to fetch weather data,
    it then updates the parent component with the weather data*/
    const handleSubmit = async (e) => {
        e.preventDefault();
        let url = 'http://localhost:8080/weather?';
        if (useCoordinates) {
            // If using coordinates, validate the inputs
            if (!latitude || !longitude) {
                setError('Please enter both latitude and longitude.');
                return;
            }
            url += `lat=${latitude}&lon=${longitude}`; // Append latitude and longitude to the URL
        } else {
            // If using city name, validate the input
            if (!location) {
                setError('Please enter a city name.');
                return;
            }
            url += `location=${location}`; // Append city name to the URL
        }

        try {
            // Make a request to the backend to fetch weather data
            const response = await fetch(url);
            const data = await response.json();
            setWeather(data); // Update the parent component with the fetched weather data
        } catch (err) {
            setError('Failed to fetch weather data.');
        }
    };

    return (
        // Form to capture user input and submit it to fetch weather data
        <form onSubmit={handleSubmit}>
            <div>

                <label>
                    <input
                        type="radio"
                        checked={!useCoordinates}
                        onChange={() => setUseCoordinates(false)}
                    />
                    City Name
                </label>
                <label>
                    <input
                        type="radio"
                        checked={useCoordinates}
                        onChange={() => setUseCoordinates(true)}
                    />
                    Coordinates
                </label>
            </div>


            {!useCoordinates ? (
                <div>
                    <input
                        type="text"
                        value={location}
                        onChange={(e) => setLocation(e.target.value)}
                        placeholder="Enter city name"
                    />
                </div>
            ) : (
                <div>
                    <input
                        type="text"
                        value={latitude}
                        onChange={(e) => setLatitude(e.target.value)}
                        placeholder="Enter latitude"
                    />
                    <input
                        type="text"
                        value={longitude}
                        onChange={(e) => setLongitude(e.target.value)}
                        placeholder="Enter longitude"
                    />
                </div>
            )}


            <button type="submit">Get Weather</button>


            {error && <p style={{ color: 'red' }}>{error}</p>}
        </form>
    );
}

export default WeatherForm;
