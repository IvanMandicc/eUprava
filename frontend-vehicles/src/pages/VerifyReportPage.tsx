import { FormEvent, useState } from 'react';
import { vehiclesApi } from '../api/vehicles';
import { ApiError } from '../api/client';
import { VerifyReportResponse } from '../types';

// Javna stranica, bez prijave — bilo ko sa kodom izveštaja može proveriti
// njegovu autentičnost (isti duh kao "Otvoreni podaci" u Angular aplikaciji).
export function VerifyReportPage() {
  const [code, setCode] = useState('');
  const [result, setResult] = useState<VerifyReportResponse | null>(null);
  const [loading, setLoading] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setResult(null);
    try {
      const res = await vehiclesApi.verifyReport(code.trim());
      setResult(res);
    } catch (err) {
      setResult({ valid: false, error: err instanceof ApiError ? err.message : 'Provera nije uspela.' });
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="card narrow">
      <h2>Provera izveštaja o vozilu</h2>
      <p className="hint">Unesite verifikacioni kod sa izveštaja da proverite njegovu autentičnost.</p>
      <form className="inline-form" onSubmit={submit}>
        <input placeholder="Verifikacioni kod" value={code} onChange={(e) => setCode(e.target.value)} required />
        <button type="submit" disabled={loading}>
          Proveri
        </button>
      </form>

      {result && !result.valid && <p className="error">{result.error ?? 'Izveštaj nije pronađen.'}</p>}

      {result?.valid && result.report && (
        <div className="card">
          <span className="badge badge-ok">Izveštaj je važeći</span>
          <p>
            <strong>
              {result.report.vehicle.make} {result.report.vehicle.model} ({result.report.vehicle.year})
            </strong>
          </p>
          <p>Tablica: {result.report.vehicle.plateNumber}</p>
          <p>VIN: {result.report.vehicle.vin}</p>
          <p>
            Status:{' '}
            <span className={`badge ${result.report.vehicle.status === 'stolen' ? 'badge-danger' : 'badge-ok'}`}>
              {result.report.vehicle.status === 'stolen' ? 'Ukradeno' : 'Registrovano'}
            </span>
          </p>
          <p>Generisano: {new Date(result.report.generatedAt).toLocaleString('sr-RS')}</p>
          <p>Broj prenosa vlasništva u istorijatu: {result.report.transfers.length}</p>
        </div>
      )}
    </div>
  );
}
