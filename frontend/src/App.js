import React, { useState } from 'react';
import axios from 'axios';

function App() {
  const [name, setName] = useState('');
  const [image, setImage] = useState('');
  const [command, setCommand] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      const response = await axios.post('/submit_job', {
        name,
        image,
        command,
      });
      alert(response.data.message);
    } catch (error) {
      alert('Error submitting job: ' + error.message);
    }
  };

  return (
    <div>
      <h1>Submit Job</h1>
      <form onSubmit={handleSubmit}>
        <div>
          <label>Name:</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
        </div>
        <div>
          <label>Image:</label>
          <input
            type="text"
            value={image}
            onChange={(e) => setImage(e.target.value)}
            required
          />
        </div>
        <div>
          <label>Command:</label>
          <input
            type="text"
            value={command}
            onChange={(e) => setCommand(e.target.value)}
            required
          />
        </div>
        <button type="submit">Submit</button>
      </form>
    </div>
  );
}

export default App;
