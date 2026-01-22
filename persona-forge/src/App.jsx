import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Hero from './components/Hero';
import ChoosePersona from './components/ChoosePersona';
import PersonaChat from './components/PersonaChat';
import './App.css';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Hero />} />
        <Route path="/personas" element={<ChoosePersona />} />
                <Route path="/personas/:id" element ={<PersonaChat />} />

      </Routes>
    </BrowserRouter>
  );
}

export default App;