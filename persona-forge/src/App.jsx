import { BrowserRouter, Routes, Route } from "react-router-dom";
import Home from "./pages/Home";
import Personas from "./pages/Personas";
import PersonaChat from "./pages/PersonaChat";
import "./App.css";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/personas" element={<Personas />} />
        <Route path="/personas/:id" element={<PersonaChat />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
