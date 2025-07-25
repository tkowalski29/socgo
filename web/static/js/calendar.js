// Calendar functionality
let currentWeek = new Date();

// Initialize calendar
document.addEventListener('DOMContentLoaded', function() {
  updateWeekTitle();
  setupEventListeners();
});

function setupEventListeners() {
  const prevWeekBtn = document.getElementById('prev-week');
  const nextWeekBtn = document.getElementById('next-week');
  const todayBtn = document.getElementById('today');

  if (prevWeekBtn) {
    prevWeekBtn.addEventListener('click', function() {
      currentWeek.setDate(currentWeek.getDate() - 7);
      updateWeekTitle();
      loadWeekView();
    });
  }

  if (nextWeekBtn) {
    nextWeekBtn.addEventListener('click', function() {
      currentWeek.setDate(currentWeek.getDate() + 7);
      updateWeekTitle();
      loadWeekView();
    });
  }

  if (todayBtn) {
    todayBtn.addEventListener('click', function() {
      currentWeek = new Date();
      updateWeekTitle();
      loadWeekView();
    });
  }
}

function updateWeekTitle() {
  const weekTitleElement = document.getElementById('week-title');
  if (!weekTitleElement) return;

  const startOfWeek = getStartOfWeek(currentWeek);
  const endOfWeek = new Date(startOfWeek);
  endOfWeek.setDate(endOfWeek.getDate() + 6);
  
  const title = `Tydzień ${startOfWeek.getDate()}-${endOfWeek.getDate()} ${getMonthName(startOfWeek.getMonth())} ${startOfWeek.getFullYear()}`;
  weekTitleElement.textContent = title;
}

function loadWeekView() {
  const startOfWeek = getStartOfWeek(currentWeek);
  const params = new URLSearchParams({
    start: startOfWeek.toISOString().split('T')[0]
  });
  
  htmx.ajax('GET', `/api/calendar/week?${params}`, {target: '#calendar-week'});
}

function getStartOfWeek(date) {
  const d = new Date(date);
  const day = d.getDay();
  const diff = d.getDate() - day + (day === 0 ? -6 : 1);
  return new Date(d.setDate(diff));
}

function getMonthName(month) {
  const months = [
    'stycznia', 'lutego', 'marca', 'kwietnia', 'maja', 'czerwca',
    'lipca', 'sierpnia', 'września', 'października', 'listopada', 'grudnia'
  ];
  return months[month];
}

function showCalendarPostSidebar(postId, postType) {
  const sidebar = document.getElementById('post-details-sidebar');
  const content = document.getElementById('sidebar-content');
  
  if (!sidebar || !content) return;
  
  // Load post details
  htmx.ajax('GET', `/api/posts/${postId}?type=${postType}`, {
    target: '#sidebar-content',
    swap: 'innerHTML'
  });
  
  // Slide in from bottom
  sidebar.classList.add('open');
}

function closeCalendarPostSidebar() {
  const sidebar = document.getElementById('post-details-sidebar');
  if (!sidebar) {
    return;
  }
  
  // Slide out to bottom
  sidebar.classList.remove('open');
}

function editPost(postId, postType) {
  // Redirect to edit page or show edit modal
  window.location.href = `/posts/edit/${postId}?type=${postType}`;
}

function deletePost(postId, postType) {
  if (confirm('Czy na pewno chcesz usunąć ten post?')) {
    htmx.ajax('DELETE', `/api/posts/${postId}?type=${postType}`, {
      target: 'body',
      swap: 'none',
      headers: {
        'HX-Request': 'true'
      }
    }).then(() => {
      loadWeekView(); // Reload calendar
      closeCalendarPostSidebar();
    });
  }
}

function toggleHourPosts(dayIndex, hour) {
  const hiddenPosts = document.getElementById(`hidden-posts-${dayIndex}-${hour}`);
  const button = document.getElementById(`toggle-${dayIndex}-${hour}`);
  
  if (!hiddenPosts || !button) return;
  
  if (hiddenPosts.classList.contains('hidden')) {
    hiddenPosts.classList.remove('hidden');
    button.textContent = 'Ukryj';
  } else {
    hiddenPosts.classList.add('hidden');
    button.textContent = `+${button.getAttribute('data-count')} więcej`;
  }
}


// Close sidebar when clicking outside
document.addEventListener('click', function(e) {
  const sidebar = document.getElementById('post-details-sidebar');
  const calendarContainer = document.getElementById('calendar-container');
  
  if (!sidebar || !calendarContainer) return;
  
  if (!sidebar.contains(e.target) && !calendarContainer.contains(e.target) && sidebar.classList.contains('open')) {
    closeCalendarPostSidebar();
  }
});

// Make functions globally available
window.showCalendarPostSidebar = showCalendarPostSidebar;
window.closeCalendarPostSidebar = closeCalendarPostSidebar;
window.editPost = editPost;
window.deletePost = deletePost;
window.toggleHourPosts = toggleHourPosts;