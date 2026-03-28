// ── Feature flags (env vars, default true) ────────────────────────────────
const FLAGS = {
  SCAN:     import.meta.env.VITE_FEATURE_SCAN     !== 'false',
  SUBMIT:   import.meta.env.VITE_FEATURE_SUBMIT   !== 'false',
  RECEIPTS: import.meta.env.VITE_FEATURE_RECEIPTS !== 'false',
  EXPORT:   import.meta.env.VITE_FEATURE_EXPORT   !== 'false',
  ACCOUNTS: import.meta.env.VITE_FEATURE_ACCOUNTS !== 'false',
  USERS:    import.meta.env.VITE_FEATURE_USERS    !== 'false',
  CLASSES:  import.meta.env.VITE_FEATURE_CLASSES  !== 'false',
  DATA:     import.meta.env.VITE_FEATURE_DATA     !== 'false',
  TEMPLATES: import.meta.env.VITE_FEATURE_TEMPLATES !== 'false',
  TIMETABLE: import.meta.env.VITE_FEATURE_TIMETABLE !== 'false',
}

// ── Feature definitions ────────────────────────────────────────────────────
// roles: any match grants access
// flag:  must be enabled in FLAGS
const FEATURES = {
  scanReceipt:     { roles: ['teacher'],           flag: 'SCAN' },
  submitReceipts:  { roles: ['teacher'],           flag: 'SUBMIT' },
  viewOwnReceipts: { roles: ['teacher'],           flag: 'RECEIPTS' },
  viewAllReceipts: { roles: ['school-management'], flag: 'RECEIPTS' },
  exportReceipts:  { roles: ['school-management'], flag: 'EXPORT' },
  manageAccounts:  { roles: ['school-management'], flag: 'ACCOUNTS' },
  manageUsers:     { roles: ['user_management'],   flag: 'USERS' },
  manageClasses:   { roles: ['user_management'],   flag: 'CLASSES' },
  manageData:      { roles: ['teacher', 'school-management'], flag: 'DATA' },
  manageTemplates: { roles: ['school-management'], flag: 'TEMPLATES' },
  timetablePlanner: { roles: ['teacher', 'school-management', 'social-worker', 'individual-instruction', 'speech-therapy', 'psychomotor-therapy'], flag: 'TIMETABLE' },
  timetableView: { roles: ['teacher', 'school-management', 'social-worker', 'individual-instruction', 'speech-therapy', 'psychomotor-therapy'], flag: 'TIMETABLE' },
}

// ── Helpers ────────────────────────────────────────────────────────────────

export function getUserRoles(user) {
  return user?.realm_access?.roles ?? []
}

// ── Domain model role helpers ──────────────────────────────────────────────

// Models where only _write_all grants the CREATE action (teachers must be admins to create).
const WRITE_ALL_CREATE_ONLY = new Set(['teacher', 'schoolclass'])
// Models where _write_own also grants CREATE.
const WRITE_OWN_CREATE_OK   = new Set(['lesson', 'exam', 'grade'])

export function canReadModel(user, modelKey) {
  return getUserRoles(user).includes(`${modelKey}_read`)
}

export function canWriteModel(user, modelKey) {
  const roles = getUserRoles(user)
  return roles.includes(`${modelKey}_write`) ||
         roles.includes(`${modelKey}_write_all`) ||
         roles.includes(`${modelKey}_write_own`)
}

export function canCreateModel(user, modelKey) {
  const roles = getUserRoles(user)
  if (WRITE_ALL_CREATE_ONLY.has(modelKey)) return roles.includes(`${modelKey}_write_all`)
  if (WRITE_OWN_CREATE_OK.has(modelKey))   return roles.includes(`${modelKey}_write_all`) || roles.includes(`${modelKey}_write_own`)
  return roles.includes(`${modelKey}_write`)
}

export function hasFeature(user, featureKey) {
  const def = FEATURES[featureKey]
  if (!def) return false
  if (!FLAGS[def.flag]) return false
  const roles = getUserRoles(user)
  return def.roles.some(r => roles.includes(r))
}
