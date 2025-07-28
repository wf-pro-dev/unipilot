"use client"

import { useAssignments } from "@/hooks/use-assignments"
import { parseDeadline, formatDeadline } from "@/lib/date-utils"

export default function AssignmentsList() {
  const { assignments, isLoading, error, refreshAssignments } = useAssignments()

  if (isLoading) {
    return <div>Loading assignments...</div>
  }

  if (error) {
    return <div>Error: {error}</div>
  }

  return (
    <div className="p-4">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-bold">Assignments ({assignments.length})</h2>
        <button 
          onClick={refreshAssignments}
          className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
        >
          Refresh
        </button>
      </div>

      {assignments.length === 0 ? (
        <p className="text-gray-500">No assignments found.</p>
      ) : (
        <div className="space-y-4">
          {assignments.map((assignment) => {
            const deadline = parseDeadline(assignment.Deadline)
            return (
              <div 
                key={assignment.ID} 
                className="border rounded-lg p-4 hover:shadow-md transition-shadow"
              >
                <h3 className="text-lg font-semibold">{assignment.Title}</h3>
                <div className="grid grid-cols-2 gap-4 mt-2 text-sm text-gray-600">
                  <div>
                    <span className="font-medium">Course:</span> {assignment.Course?.Code || 'N/A'}
                  </div>
                  <div>
                    <span className="font-medium">Type:</span> {assignment.TypeName}
                  </div>
                  <div>
                    <span className="font-medium">Status:</span> {assignment.StatusName}
                  </div>
                  <div>
                    <span className="font-medium">Due:</span> {formatDeadline(deadline)}
                  </div>
                </div>
                {assignment.Todo && (
                  <p className="mt-2 text-gray-700">{assignment.Todo}</p>
                )}
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}