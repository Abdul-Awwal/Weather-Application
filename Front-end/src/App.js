import React from 'react';
import WeatherForm from './WeatherForm';
import WeatherDisplay from './WeatherDisplay';

function App() {
  const [weather, setWeather] = React.useState(null);

  return (
      <div className="App">
        <h1>Weather App</h1>
        <WeatherForm setWeather={setWeather} />
        {weather && <WeatherDisplay weather={weather} />}
      </div>
  );
}

export default App;
