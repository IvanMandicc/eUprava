import { FormEvent, useEffect, useState } from 'react';
import { vehiclesApi } from '../../api/vehicles';
import { ApiError } from '../../api/client';
import { PlateReservation } from '../../types';

const statusLabel: Record<PlateReservation['status'], string> = {
  pending: 'Na čekanju',
  approved: 'Odobreno',
  rejected: 'Odbijeno',
  expired: 'Isteklo',
};

const statusClass: Record<PlateReservation['status'], string> = {
  pending: 'badge-warn',
  approved: 'badge-ok',
  rejected: 'badge-danger',
  expired: 'badge-danger',
};

export function PlatesPage() {
  const [plate, setPlate] = useState('');
  const [reservations, setReservations] = useState<PlateReservation[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  function load() {
    vehiclesApi
      .myReservations()
      .then(setReservations)
      .catch((err) => setError(err instanceof ApiError ? err.message : 'Učitavanje nije uspelo.'));
  }

  useEffect(load, []);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      await vehiclesApi.requestPlate(plate);
      setPlate('');
      load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Zahtev nije uspeo.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <div className="card">
        <h2>Zahtev za personalizovanu tablicu</h2>
        <form className="inline-form" onSubmit={submit}>
          <input
            placeholder="npr. MARKO01"
            value={plate}
            onChange={(e) => setPlate(e.target.value.toUpperCase())}
            required
          />
          <button type="submit" disabled={loading}>
            Pošalji zahtev
          </button>
        </form>
        {error && <p className="error">{error}</p>}
        <p className="hint">4–7 znakova (slova A-Z bez Q, W, X, Y i cifre). Kombinacije do 5 znakova nose višu taksu.</p>
      </div>

      <div className="card">
        <h3>Moji zahtevi</h3>
        {reservations.length === 0 ? (
          <p className="hint">Nemate podnetih zahteva.</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Tablica</th>
                <th>Taksa (RSD)</th>
                <th>Podneto</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {reservations.map((r) => (
                <tr key={r.id}>
                  <td>{r.requestedPlate}</td>
                  <td>{r.feeAmount.toFixed(2)}</td>
                  <td>{new Date(r.requestedAt).toLocaleDateString('sr-RS')}</td>
                  <td>
                    <span className={`badge ${statusClass[r.status]}`}>{statusLabel[r.status]}</span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </>
  );
}
