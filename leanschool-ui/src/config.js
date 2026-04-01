const w = (typeof window !== 'undefined' && window.__LS_CONFIG__) || {}
const env = import.meta.env

export const config = {
  keycloakUrl:          w.keycloakUrl          || env.VITE_KEYCLOAK_URL           || 'http://localhost:8180',
  keycloakRealm:        w.keycloakRealm        || env.VITE_KEYCLOAK_REALM         || 'leanschool',
  keycloakClientId:     w.keycloakClientId     || env.VITE_KEYCLOAK_CLIENT_ID     || 'leanschool-ui',
  leanschoolUrl:        w.leanschoolUrl        || env.VITE_LEANSCHOOL_URL         || 'http://localhost:8080',
  fileServiceUrl:       w.fileServiceUrl       || env.VITE_FILE_SERVICE_URL       || 'http://localhost:8083',
  receiptReaderUrl:     w.receiptReaderUrl     || env.VITE_RECEIPT_READER_URL     || 'http://localhost:8081',
  extractionServiceUrl: w.extractionServiceUrl || env.VITE_EXTRACTION_SERVICE_URL || 'http://localhost:8084',
  timetableServiceUrl:  w.timetableServiceUrl  || env.VITE_TIMETABLE_SERVICE_URL  || 'http://localhost:8085',
  features: {
    scan:      (w.features?.scan      || env.VITE_FEATURE_SCAN)      !== 'false',
    submit:    (w.features?.submit    || env.VITE_FEATURE_SUBMIT)    !== 'false',
    receipts:  (w.features?.receipts  || env.VITE_FEATURE_RECEIPTS)  !== 'false',
    export:    (w.features?.export    || env.VITE_FEATURE_EXPORT)    !== 'false',
    accounts:  (w.features?.accounts  || env.VITE_FEATURE_ACCOUNTS)  !== 'false',
    users:     (w.features?.users     || env.VITE_FEATURE_USERS)     !== 'false',
    classes:   (w.features?.classes   || env.VITE_FEATURE_CLASSES)   !== 'false',
    data:      (w.features?.data      || env.VITE_FEATURE_DATA)      !== 'false',
    templates: (w.features?.templates || env.VITE_FEATURE_TEMPLATES) !== 'false',
    timetable: (w.features?.timetable || env.VITE_FEATURE_TIMETABLE) !== 'false',
  },
}
