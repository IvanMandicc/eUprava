import { FormEvent, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';
import { ApiError } from '../api/client';

export function LoginPage() {
  const auth = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      const user = await auth.login(email, password);
      navigate(user.role === 'officer' ? '/officer/vehicles' : '/citizen/vehicles');
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Prijava nije uspela.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="card narrow">
      <h2>Prijava</h2>
      <form onSubmit={submit}>
        <label>Email</label>
        <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        <label>Lozinka</label>
        <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
        {error && <p className="error">{error}</p>}
        <button type="submit" disabled={loading}>
          Prijavi se
        </button>
      </form>
      <p className="hint">
        Nemate nalog? <Link to="/register">Registrujte se</Link>
      </p>
    </div>
  );
}
