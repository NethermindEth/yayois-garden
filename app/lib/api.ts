export async function getLastUpdate() {
  try {
    return await mockGetLastUpdate()
  } catch (error) {
    console.error('Error fetching last update:', error)
    return {
      date: 'N/A',
      message: 'Unable to fetch update'
    }
  }
}

// Simulated API response for demonstration
async function mockGetLastUpdate() {
  await new Promise(resolve => setTimeout(resolve, 100)) // Simulate network delay
  const date = new Date().toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })
  return {
    date,
    message: '>_ AI art... is art'
  }
}

