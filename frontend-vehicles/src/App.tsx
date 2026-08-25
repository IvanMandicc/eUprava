import { Navigate, Route, Routes } from 'react-router-dom';
import { Layout } from './components/Layout';
import { RequireRole } from './components/RequireRole';
import { LoginPage } from './pages/LoginPage';
import { RegisterPage } from './pages/RegisterPage';
import { VerifyReportPage } from './pages/VerifyReportPage';
import { MyVehiclesPage } from './pages/citizen/MyVehiclesPage';
import { PlatesPage } from './pages/citizen/PlatesPage';
import { VehiclesPage } from './pages/officer/VehiclesPage';
import { PlateRequestsPage } from './pages/officer/PlateRequestsPage';

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<Navigate to="/login" replace />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/verify" element={<VerifyReportPage />} />

        <Route
          path="/citizen/vehicles"
          element={
            <RequireRole roles={['citizen']}>
              <MyVehiclesPage />
            </RequireRole>
          }
        />
        <Route
          path="/citizen/plates"
          element={
            <RequireRole roles={['citizen']}>
              <PlatesPage />
            </RequireRole>
          }
        />

        <Route
          path="/officer/vehicles"
          element={
            <RequireRole roles={['officer']}>
              <VehiclesPage />
            </RequireRole>
          }
        />
        <Route
          path="/officer/plate-requests"
          element={
            <RequireRole roles={['officer']}>
              <PlateRequestsPage />
            </RequireRole>
          }
        />

        <Route path="*" element={<Navigate to="/login" replace />} />
      </Route>
    </Routes>
  );
}
