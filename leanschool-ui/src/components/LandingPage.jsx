import { useTranslation } from '../i18n/useTranslation'
import './LandingPage.css'

export default function LandingPage({ onLogin }) {
  const { t } = useTranslation()

  return (
    <div className="landing">
      {/* animated background orbs */}
      <div className="orb orb-1" />
      <div className="orb orb-2" />
      <div className="orb orb-3" />

      {/* nav */}
      <nav className="landing-nav">
        <img src="/logo.svg" alt={t.app.name} className="nav-logo-img" />
      </nav>

      {/* hero */}
      <section className="hero">
        <h1 className="hero-title">
          {t.landing.hero.title.split('\n').map((line, i) => (
            <span key={i} className={i === 1 ? 'gradient-text' : ''}>{line}<br /></span>
          ))}
        </h1>
        <p className="hero-subtitle">{t.landing.hero.subtitle}</p>

        <div className="hero-actions">
          <button className="cta-button" onClick={onLogin}>
            {t.login.signIn}
          </button>
          {/* Register redirects to Keycloak login; when registrationAllowed=true
              Keycloak shows a "Register" link on its login page. */}
          <button className="ghost-button" onClick={onLogin}>
            {t.login.register}
          </button>
        </div>


      </section>

      {/* features */}
      <section className="features">
        {Object.values(t.landing.features).map((f, i) => (
          <div key={i} className="feature-card">
            <div className="feature-icon">{['⬡', '✦', '◈'][i]}</div>
            <h3>{f.title}</h3>
            <p>{f.body}</p>
          </div>
        ))}
      </section>

      <footer className="landing-footer">
        <span>{t.app.name} · {t.app.tagline} · initiated by <a href="https://xn--hbu-qla.ch">häbu.ch</a></span>
      </footer>
    </div>
  )
}
