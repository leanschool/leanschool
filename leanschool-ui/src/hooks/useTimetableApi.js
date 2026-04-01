import { useAuth } from '../auth/useAuth'
import { config } from '../config'

const API = config.timetableServiceUrl

export function useTimetableApi() {
  const { authFetch } = useAuth()

  async function request(path, options = {}) {
    const res = await authFetch(`${API}${path}`, options)
    if (!res.ok) {
      const text = await res.text().catch(() => '')
      const err = new Error(text || `HTTP ${res.status}`)
      err.status = res.status
      throw err
    }
    if (res.status === 204) return null
    return res.json()
  }

  function post(path, body) {
    return request(path, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
  }
  function put(path, body) {
    return request(path, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
  }
  function del(path) {
    return request(path, { method: 'DELETE' })
  }

  return {
    // Plans
    listPlans: () => request('/plans'),
    getPlan: (id) => request(`/plans/${id}`),
    createPlan: (data) => post('/plans', data),
    updatePlan: (id, data) => put(`/plans/${id}`, data),
    deletePlan: (id) => del(`/plans/${id}`),

    // Time slots
    listTimeSlots: (planId) => request(`/plans/${planId}/time-slots`),
    createTimeSlot: (planId, data) => post(`/plans/${planId}/time-slots`, data),
    updateTimeSlot: (planId, id, data) => put(`/plans/${planId}/time-slots/${id}`, data),
    deleteTimeSlot: (planId, id) => del(`/plans/${planId}/time-slots/${id}`),
    generateDefaultSlots: (planId, data) => post(`/plans/${planId}/time-slots/generate-default`, data),

    // Requirements
    listRequirements: (planId) => request(`/plans/${planId}/requirements`),
    createRequirement: (planId, data) => post(`/plans/${planId}/requirements`, data),
    updateRequirement: (planId, id, data) => put(`/plans/${planId}/requirements/${id}`, data),
    deleteRequirement: (planId, id) => del(`/plans/${planId}/requirements/${id}`),

    // Constraints
    listConstraints: (planId) => request(`/plans/${planId}/constraints`),
    createConstraint: (planId, data) => post(`/plans/${planId}/constraints`, data),
    updateConstraint: (planId, id, data) => put(`/plans/${planId}/constraints/${id}`, data),
    deleteConstraint: (planId, id) => del(`/plans/${planId}/constraints/${id}`),

    // Entries
    listEntries: (planId, filters = {}) => {
      const params = new URLSearchParams()
      if (filters.classId) params.set('classId', filters.classId)
      if (filters.teacherId) params.set('teacherId', filters.teacherId)
      const qs = params.toString()
      return request(`/plans/${planId}/entries${qs ? '?' + qs : ''}`)
    },
    getEntry: (planId, id) => request(`/plans/${planId}/entries/${id}`),
    updateEntry: (planId, id, data) => put(`/plans/${planId}/entries/${id}`, data),
    swapEntries: (planId, id, targetEntryId) => post(`/plans/${planId}/entries/${id}/swap`, { targetEntryId }),
    reassignTeacher: (planId, id, teacherId) => post(`/plans/${planId}/entries/${id}/reassign`, { teacherId }),

    // Conflicts
    listConflicts: (planId, filters = {}) => {
      const params = new URLSearchParams()
      if (filters.resolved !== undefined) params.set('resolved', String(filters.resolved))
      if (filters.teacherId) params.set('teacherId', filters.teacherId)
      const qs = params.toString()
      return request(`/plans/${planId}/conflicts${qs ? '?' + qs : ''}`)
    },

    // Snapshots (read-only)
    listTeachers: (planId) => request(`/plans/${planId}/teachers`),
    listSubjects: (planId) => request(`/plans/${planId}/subjects`),
    listClasses: (planId) => request(`/plans/${planId}/classes`),
    listRooms: (planId) => request(`/plans/${planId}/rooms`),

    // Workflow
    takeSnapshot: (planId) => post(`/plans/${planId}/snapshot`),
    generate: (planId) => post(`/plans/${planId}/generate`),
    validate: (planId) => post(`/plans/${planId}/validate`),
    finalize: (planId) => post(`/plans/${planId}/finalize`),
    reset: (planId) => post(`/plans/${planId}/reset`),
  }
}
