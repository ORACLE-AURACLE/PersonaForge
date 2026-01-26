import { useNavigate } from "react-router-dom"; // Importing from the parent directory
import logo from "../assets/images/Main-Logo.svg";
import rectImg from "../assets/images/Rectangle.svg";

export default function Home() {
  const navigate = useNavigate();
  const bars = Array.from({ length: 9 });

  return (
    <section className="hero-container">
      {/* Top Group */}
      <div className="decor-group top-right">
        {bars.map((_, i) => (
          <img
            key={`t-${i}`}
            src={rectImg}
            className={`rect-bar bar-${i + 1}`}
            alt=""
          />
        ))}
      </div>

      <nav className="nav-wrapper">
        <img src={logo} alt="PersonaForge" className="logo" />
      </nav>

      <div className="heroContent">
        <h1>Talk to real people before they exist</h1>
        <p>
          Choose a realistic persona, test your idea, and understand why they
          react the way they do.
        </p>
        <button className="cta-button" onClick={() => navigate("/personas")}>
          Choose a Persona
        </button>
      </div>

      {/* Bottom Group */}
      <div className="decor-group bottom-left">
        {bars.map((_, i) => (
          <img
            key={`b-${i}`}
            src={rectImg}
            className={`rect-bar bar-${i + 1}`}
            alt=""
          />
        ))}
      </div>
    </section>
  );
}
