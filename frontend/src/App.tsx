import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Layout from './components/layout/Layout';
import HomePage from './pages/HomePage';
import RecipeListPage from './pages/RecipeListPage';
import WeeklyPlanPage from './pages/WeeklyPlanPage';
import SettingsPage from './pages/SettingsPage';

function App() {
  return (
    <Router>
      <Layout>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/recipes" element={<RecipeListPage />} />
          <Route path="/meal-plan" element={<WeeklyPlanPage />} />
          <Route path="/settings" element={<SettingsPage />} />
        </Routes>
      </Layout>
    </Router>
  )
}

export default App
