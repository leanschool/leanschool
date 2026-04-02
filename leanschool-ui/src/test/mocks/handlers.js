import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('http://localhost:8080/users/me', () =>
    HttpResponse.json({ registrationStatus: 'none' })
  ),
  http.get('http://localhost:8080/users/role-mappings', () =>
    HttpResponse.json([
      { name: 'teacher', description: 'Teacher' },
      { name: 'school-management', description: 'School Management' },
    ])
  ),
  http.get('http://localhost:8080/registration/school-classes', () =>
    HttpResponse.json([
      { id: 'c1', name: '1a' },
      { id: 'c2', name: '2b' },
    ])
  ),
  http.post('http://localhost:8080/registration/start', () =>
    new HttpResponse(null, { status: 200 })
  ),
]
