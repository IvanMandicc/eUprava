import { FormEvent, useEffect, useRef, useState } from 'react';
import { RegisterVehicleInput, vehiclesApi } from '../../api/vehicles';
import { citizensApi } from '../../api/citizens';
import { ApiError } from '../../api/client';
import { User, Vehicle } from '../../types';

const emptyInput: RegisterVehicleInput = {
  ownerCitizenId: 0,
  vin: '',
  make: '',
  model: '',
  year: new Date().getFullYear(),
  category: 'M1',
  color: '',
  enginePowerKw: 0,
  fuelType: 'benzin',
  insuranceValidUntil: '',
  techInspectionValidUntil: '',
};

export function VehiclesPage() {
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [input, setInput] = useState<RegisterVehicleInput>(emptyInput);
  const [newOwnerByVehicle, setNewOwnerByVehicle] = useState<Record<number, string>>({});
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [busyId, setBusyId] = useState<number | null>(null);

  const [searchJmbg, setSearchJmbg] = useState('');
  const [foundCitizen, setFoundCitizen] = useState<User | null>(null);
  const [searchError, setSearchError] = useState('');
  const [suggestions, setSuggestions] = useState<User[]>([]);
  const suggestTimer = useRef<ReturnType<typeof setTimeout>>();

  function load() {
    vehiclesApi
      .list()
      .then(setVehicles)
      .catch((err) => setError(err instanceof ApiError ? err.message : 'Učitavanje nije uspelo.'));
  }

  useEffect(load, []);

  function set<K extends keyof RegisterVehicleInput>(key: K, value: RegisterVehicleInput[K]) {
    setInput((prev) => ({ ...prev, [key]: value }));
  }

  function searchCitizen(e: FormEvent) {
    e.preventDefault();
    setSearchError('');
    setFoundCitizen(null);
    citizensApi
      .searchByJmbg(searchJmbg)
      .then((u) => {
        setFoundCitizen(u);
        set('ownerCitizenId', u.id);
      })
      .catch((err) => setSearchError(err instanceof ApiError ? err.message : 'Građanin nije pronađen.'));
  }

  // Autocomplete: dok policajac kuca, sa malim zakašnjenjem (debounce) šalje
  // se zahtev za predlozima — ne na svaki taster, već tek kad kucanje stane.
  function onJmbgInput(value: string) {
    setSearchJmbg(value);
    setFoundCitizen(null);
    setSearchError('');
    clearTimeout(suggestTimer.current);
    if (value.length < 3) {
      setSuggestions([]);
      return;
    }
    suggestTimer.current = setTimeout(() => {
      citizensApi
        .suggestByJmbg(value)
        .then(setSuggestions)
        .catch(() => setSuggestions([]));
    }, 300);
  }

  function pickSuggestion(u: User) {
    setFoundCitizen(u);
    setSearchJmbg(u.jmbg);
    set('ownerCitizenId', u.id);
    setSuggestions([]);
  }

  // mousedown na stavci liste izvrši pickSuggestion PRE nego što input
  // izgubi fokus (blur) — zato ovde samo sakrivamo listu sa malim kašnjenjem,
  // da klik stigne da se registruje.
  function hideSuggestionsSoon() {
    setTimeout(() => setSuggestions([]), 150);
  }

  async function submit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      await vehiclesApi.register({
        ...input,
        insuranceValidUntil: new Date(input.insuranceValidUntil).toISOString(),
        techInspectionValidUntil: new Date(input.techInspectionValidUntil).toISOString(),
      });
      setInput(emptyInput);
      load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Registracija vozila nije uspela.');
    } finally {
      setLoading(false);
    }
  }

  async function transfer(id: number) {
    const newOwnerCitizenId = Number(newOwnerByVehicle[id]);
    if (!newOwnerCitizenId) return;
    setBusyId(id);
    setError('');
    try {
      await vehiclesApi.transfer(id, newOwnerCitizenId);
      load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Prenos vlasništva nije uspeo.');
    } finally {
      setBusyId(null);
    }
  }

  async function renew(id: number) {
    setBusyId(id);
    setError('');
    try {
      await vehiclesApi.renew(id);
      load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Produženje registracije nije uspelo.');
    } finally {
      setBusyId(null);
    }
  }

  return (
    <>
      <div className="card">
        <h2>Registracija vozila</h2>

        <p className="hint">Pronađi vlasnika po JMBG-u, pa registruj vozilo na njegovo ime.</p>
        <form className="inline-form" onSubmit={searchCitizen}>
          <div className="autocomplete-wrapper">
            <input
              value={searchJmbg}
              onChange={(e) => onJmbgInput(e.target.value)}
              onBlur={hideSuggestionsSoon}
              placeholder="JMBG građanina (kucaj — predlozi se pojavljuju usput)"
              autoComplete="off"
              required
              minLength={13}
              maxLength={13}
            />
            {suggestions.length > 0 && (
              <ul className="autocomplete-list">
                {suggestions.map((u) => (
                  <li key={u.id} onMouseDown={() => pickSuggestion(u)}>
                    <strong>{u.jmbg}</strong> — {u.firstName} {u.lastName} ({u.email})
                  </li>
                ))}
              </ul>
            )}
          </div>
          <button type="submit">Pretraži</button>
        </form>
        {searchError && <p className="error">{searchError}</p>}
        {foundCitizen && (
          <p className="hint">
            Pronađen: <strong>{foundCitizen.firstName} {foundCitizen.lastName}</strong> ({foundCitizen.email})
          </p>
        )}

        <form onSubmit={submit}>
          <label>ID građanina (vlasnik)</label>
          <input
            type="number"
            value={input.ownerCitizenId || ''}
            onChange={(e) => set('ownerCitizenId', Number(e.target.value))}
            required
          />
          <label>Broj šasije (VIN, 17 znakova)</label>
          <input value={input.vin} onChange={(e) => set('vin', e.target.value.toUpperCase())} required maxLength={17} />
          <div className="inline-form">
            <div>
              <label>Marka</label>
              <input value={input.make} onChange={(e) => set('make', e.target.value)} required />
            </div>
            <div>
              <label>Model</label>
              <input value={input.model} onChange={(e) => set('model', e.target.value)} required />
            </div>
            <div>
              <label>Godište</label>
              <input type="number" value={input.year} onChange={(e) => set('year', Number(e.target.value))} required />
            </div>
          </div>
          <div className="inline-form">
            <div>
              <label>Kategorija</label>
              <select value={input.category} onChange={(e) => set('category', e.target.value)}>
                {['M1', 'M2', 'M3', 'N1', 'N2', 'N3', 'L'].map((c) => (
                  <option key={c} value={c}>
                    {c}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label>Gorivo</label>
              <select value={input.fuelType} onChange={(e) => set('fuelType', e.target.value)}>
                {['benzin', 'dizel', 'hibrid', 'elektro', 'gas'].map((f) => (
                  <option key={f} value={f}>
                    {f}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label>Snaga motora (kW)</label>
              <input
                type="number"
                value={input.enginePowerKw}
                onChange={(e) => set('enginePowerKw', Number(e.target.value))}
              />
            </div>
          </div>
          <div className="inline-form">
            <div>
              <label>Boja</label>
              <input value={input.color} onChange={(e) => set('color', e.target.value)} />
            </div>
            <div>
              <label>Osiguranje važi do</label>
              <input
                type="date"
                value={input.insuranceValidUntil}
                onChange={(e) => set('insuranceValidUntil', e.target.value)}
                required
              />
            </div>
            <div>
              <label>Tehnički pregled važi do</label>
              <input
                type="date"
                value={input.techInspectionValidUntil}
                onChange={(e) => set('techInspectionValidUntil', e.target.value)}
                required
              />
            </div>
          </div>
          {error && <p className="error">{error}</p>}
          <button type="submit" disabled={loading}>
            Registruj vozilo
          </button>
        </form>
      </div>

      <div className="card">
        <h3>Registar vozila</h3>
        {vehicles.length === 0 ? (
          <p className="hint">Nema registrovanih vozila.</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Tablica</th>
                <th>Vozilo</th>
                <th>Vlasnik (ID)</th>
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
                  <td>{v.ownerCitizenId}</td>
                  <td>{new Date(v.registrationValidUntil).toLocaleDateString('sr-RS')}</td>
                  <td>
                    <span className={`badge ${v.status === 'stolen' ? 'badge-danger' : 'badge-ok'}`}>
                      {v.status === 'stolen' ? 'Ukradeno' : 'Registrovano'}
                    </span>
                  </td>
                  <td className="actions">
                    <input
                      type="number"
                      placeholder="ID novog vlasnika"
                      style={{ width: 130, display: 'inline-block' }}
                      value={newOwnerByVehicle[v.id] ?? ''}
                      onChange={(e) => setNewOwnerByVehicle((prev) => ({ ...prev, [v.id]: e.target.value }))}
                    />
                    <button className="secondary" disabled={busyId === v.id} onClick={() => transfer(v.id)}>
                      Prenesi
                    </button>
                    <button className="secondary" disabled={busyId === v.id} onClick={() => renew(v.id)}>
                      Produži registraciju
                    </button>
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
