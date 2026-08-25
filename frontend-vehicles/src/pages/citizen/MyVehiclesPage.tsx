import { useEffect, useState } from 'react';
import { vehiclesApi } from '../../api/vehicles';
import { ApiError } from '../../api/client';
import { Vehicle } from '../../types';

export function MyVehiclesPage() {
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [error, setError] = useState('');
  const [busyId, setBusyId] = useState<number | null>(null);
  const [reportCodes, setReportCodes] = useState<Record<number, string>>({});

  function load() {
    vehiclesApi
      .myVehicles()
      .then(setVehicles)
      .catch((err) => setError(err instanceof ApiError ? err.message : 'Učitavanje nije uspelo.'));
  }

  useEffect(load, []);

  async function reportTheft(id: number) {
    setBusyId(id);
    setError('');
    try {
      await vehiclesApi.reportTheft(id);
      load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Prijava krađe nije uspela.');
    } finally {
      setBusyId(null);
    }
  }

  async function reportFound(id: number) {
    setBusyId(id);
    setError('');
    try {
      await vehiclesApi.reportFound(id);
      load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Prijava pronalaska nije uspela.');
    } finally {
      setBusyId(null);
    }
  }

  async function generateReport(id: number) {
    setBusyId(id);
    setError('');
    try {
      const report = await vehiclesApi.generateReport(id);
      setReportCodes((prev) => ({ ...prev, [id]: report.verificationCode }));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Generisanje izveštaja nije uspelo.');
    } finally {
      setBusyId(null);
    }
  }

  return (
    <div className="card">
      <h2>Moja vozila</h2>
      {error && <p className="error">{error}</p>}
      {vehicles.length === 0 ? (
        <p className="hint">Nemate registrovanih vozila.</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Tablica</th>
              <th>Vozilo</th>
              <th>Registracija do</th>
              <th>Status</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {vehicles.map((v) => (
              <tr key={v.id}>
                <td>{v.plateNumber}</td>
                <td>
                  {v.make} {v.model} ({v.year})
                </td>
                <td>{new Date(v.registrationValidUntil).toLocaleDateString('sr-RS')}</td>
                <td>
                  <span className={`badge ${v.status === 'stolen' ? 'badge-danger' : 'badge-ok'}`}>
                    {v.status === 'stolen' ? 'Ukradeno' : 'Registrovano'}
                  </span>
                </td>
                <td className="actions">
                  {v.status === 'stolen' ? (
                    <button className="secondary" disabled={busyId === v.id} onClick={() => reportFound(v.id)}>
                      Pronađeno
                    </button>
                  ) : (
                    <button className="danger" disabled={busyId === v.id} onClick={() => reportTheft(v.id)}>
                      Prijavi krađu
                    </button>
                  )}
                  <button className="secondary" disabled={busyId === v.id} onClick={() => generateReport(v.id)}>
                    Generiši izveštaj
                  </button>
                  {reportCodes[v.id] && (
                    <p className="hint">
                      Kod za verifikaciju: <strong>{reportCodes[v.id]}</strong>
                    </p>
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
