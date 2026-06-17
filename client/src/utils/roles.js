export function roleDashboard(role) {
  const map = {
    university: { name: 'UniversityDashboard' },
    company:    { name: 'CompanyDashboard' },
    student:    { name: 'StudentDashboard' },
  }
  return map[role] ?? { name: 'Home' }
}