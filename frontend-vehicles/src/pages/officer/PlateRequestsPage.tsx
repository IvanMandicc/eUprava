import { useEffect, useState } from 'react';
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

export function PlateRequestsPage() {
  const [reservations, setReservations] = useState<PlateReservation[]>([]);
  const [error, setError] = useState('');
  const [busyId, setBusyId] = useState<number | null>(null);

  function load() {
    vehiclesApi
      .listReservations()
      .then(setReservations)
      .catch((err) => setError(err instanceof ApiError ? err.message : 'Učitavanje nije uspelo.'));
  }

  useEffect(load, []);

  async function decide(id: number, approve: boolean) {
    setBusyId(id);
    setError('');
    try {
      await vehiclesApi.decideReservation(id, approve);
      load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Obrada zahteva nije uspela.');
    } finally {
      setBusyId(null);
    }
  }

  return (
    <div className="card">
      <h2>Zahtevi za personalizovane tablice</h2>
      {error && <p className="error">{error}</p>}
      {reservations.length === 0 ? (
        <p className="hint">Nema podnetih zahteva.</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Tablica</th>
              <th>Podnosilac (ID)</th>
              <th>Taksa (RSD)</th>
              <th>Podneto</th>
              <th>Status</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {reservations.map((r) => (
              <tr key={r.id}>
                <td>{r.requestedPlate}</td>
                <td>{r.requestedByCitizenId}</td>
                <td>{r.feeAmount.toFixed(2)}</td>
                <td>{new Date(r.requestedAt).toLocaleDateString('sr-RS')}</td>
                <td>
                  <span className={`badge ${statusClass[r.status]}`}>{statusLabel[r.status]}</span>
                </td>
                <td className="actions">
                  {r.status === 'pending' && (
                    <>
                      <button disabled={busyId === r.id} onClick={() => decide(r.id, true)}>
                        Odobri
                      </button>
                      <button className="danger" disabled={busyId === r.id} onClick={() => decide(r.id, false)}>
                        Odbij
                      </button>
                    </>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
