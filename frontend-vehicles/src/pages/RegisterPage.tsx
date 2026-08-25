import { FormEvent, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { authApi, RegisterInput } from '../api/auth';
import { ApiError } from '../api/client';

const empty: RegisterInput = { jmbg: '', firstName: '', lastName: '', email: '', password: '', address: '' };

export function RegisterPage() {
  const navigate = useNavigate();
  const [input, setInput] = useState<RegisterInput>(empty);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  function set<K extends keyof RegisterInput>(key: K, value: RegisterInput[K]) {
    setInput((prev) => ({ ...prev, [key]: value }));
  }

  async function submit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      await authApi.register(input);
      navigate('/login');
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Registracija nije uspela.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="card narrow">
      <h2>Registracija građanina</h2>
      <form onSubmit={submit}>
        <label>JMBG</label>
        <input value={input.jmbg} onChange={(e) => set('jmbg', e.target.value)} required minLength={13} maxLength={13} />
        <label>Ime</label>
        <input value={input.firstName} onChange={(e) => set('firstName', e.target.value)} required />
        <label>Prezime</label>
        <input value={input.lastName} onChange={(e) => set('lastName', e.target.value)} required />
        <label>Email</label>
        <input type="email" value={input.email} onChange={(e) => set('email', e.target.value)} required />
        <label>Lozinka (min 6 karaktera)</label>
        <input
          type="password"
          value={input.password}
          onChange={(e) => set('password', e.target.value)}
          required
          minLength={6}
        />
        <label>Adresa</label>
        <input value={input.address} onChange={(e) => set('address', e.target.value)} />
        {error && <p className="error">{error}</p>}
        <button type="submit" disabled={loading}>
          Registruj se
        </button>
      </form>
      <p className="hint">
        Već imate nalog? <Link to="/login">Prijavite se</Link>
      </p>
    </div>
  );
}
