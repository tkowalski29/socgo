// Notification system JavaScript

let currentNotificationTab = 'all';
let currentNotificationGroup = null;

// Initialize notifications
function initializeNotifications() {
    loadNotificationBell();
    loadNotificationSidebar();
}

// Load notification bell with stats
function loadNotificationBell() {
    fetch('/api/notifications/bell')
        .then(response => response.text())
        .then(html => {
            const container = document.getElementById('notification-bell-container');
            if (container) {
                container.innerHTML = html;
            }
        })
        .catch(error => {
            console.error('Error loading notification bell:', error);
        });
}

// Load notification sidebar
function loadNotificationSidebar() {
    const sidebar = document.getElementById('notification-sidebar');
    if (!sidebar) {
        const sidebarHTML = `
            <div id="notification-sidebar" class="fixed top-0 right-0 h-full w-96 bg-white shadow-lg transform translate-x-full transition-transform duration-300 z-50">
                <div class="h-full flex flex-col">
                    <!-- Header -->
                    <div class="p-4 border-b border-gray-200 bg-gray-50">
                        <div class="flex justify-between items-center">
                            <h3 class="text-lg font-semibold text-gray-900">Powiadomienia</h3>
                            <button 
                                onclick="closeNotificationSidebar()"
                                class="text-gray-400 hover:text-gray-600 transition-colors"
                            >
                                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                                </svg>
                            </button>
                        </div>
                    </div>

                    <!-- Tabs -->
                    <div class="border-b border-gray-200">
                        <div class="flex">
                            <button 
                                id="tab-all"
                                onclick="switchNotificationTab('all')"
                                class="flex-1 px-4 py-2 text-sm font-medium text-center border-b-2 border-blue-500 text-blue-600"
                            >
                                Wszystkie
                            </button>
                            <button 
                                id="tab-app"
                                onclick="switchNotificationTab('app')"
                                class="flex-1 px-4 py-2 text-sm font-medium text-center border-b-2 border-transparent text-gray-500 hover:text-gray-700"
                            >
                                Aplikacja
                            </button>
                            <button 
                                id="tab-post"
                                onclick="switchNotificationTab('post')"
                                class="flex-1 px-4 py-2 text-sm font-medium text-center border-b-2 border-transparent text-gray-500 hover:text-gray-700"
                            >
                                Posty
                            </button>
                            <button 
                                id="tab-schedule"
                                onclick="switchNotificationTab('schedule')"
                                class="flex-1 px-4 py-2 text-sm font-medium text-center border-b-2 border-transparent text-gray-500 hover:text-gray-700"
                            >
                                Harmonogram
                            </button>
                        </div>
                    </div>

                    <!-- Content -->
                    <div id="notification-content" class="flex-1 overflow-y-auto p-4">
                        <div class="flex items-center justify-center py-8">
                            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                            <span class="ml-2 text-gray-600">Ładowanie powiadomień...</span>
                        </div>
                    </div>
                </div>
            </div>
        `;
        document.body.insertAdjacentHTML('beforeend', sidebarHTML);
    }
}

// Open notification sidebar
function openNotificationSidebar() {
    const sidebar = document.getElementById('notification-sidebar');
    if (sidebar) {
        sidebar.classList.remove('translate-x-full');
        loadNotificationGroups(currentNotificationTab);
    }
}

// Close notification sidebar
function closeNotificationSidebar() {
    const sidebar = document.getElementById('notification-sidebar');
    if (sidebar) {
        sidebar.classList.add('translate-x-full');
        currentNotificationGroup = null;
    }
}

// Switch notification tab
function switchNotificationTab(tab) {
    currentNotificationTab = tab;
    currentNotificationGroup = null;
    
    // Update tab styles
    document.querySelectorAll('[id^="tab-"]').forEach(button => {
        button.classList.remove('border-blue-500', 'text-blue-600');
        button.classList.add('border-transparent', 'text-gray-500');
    });
    
    const activeTab = document.getElementById(`tab-${tab}`);
    if (activeTab) {
        activeTab.classList.remove('border-transparent', 'text-gray-500');
        activeTab.classList.add('border-blue-500', 'text-blue-600');
    }
    
    loadNotificationGroups(tab);
}

// Load notification groups
function loadNotificationGroups(type) {
    const content = document.getElementById('notification-content');
    if (!content) return;
    
    content.innerHTML = `
        <div class="flex items-center justify-center py-8">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            <span class="ml-2 text-gray-600">Ładowanie powiadomień...</span>
        </div>
    `;
    
    const url = type === 'all' ? '/api/notifications/groups-ui' : `/api/notifications/groups-ui?type=${type}`;
    
    fetch(url)
        .then(response => response.text())
        .then(html => {
            content.innerHTML = html;
        })
        .catch(error => {
            console.error('Error loading notification groups:', error);
            content.innerHTML = `
                <div class="flex items-center justify-center py-8 text-red-500">
                    <p>Błąd ładowania powiadomień</p>
                </div>
            `;
        });
}

// Open notification group
function openNotificationGroup(groupId) {
    currentNotificationGroup = groupId;
    
    const content = document.getElementById('notification-content');
    if (!content) return;
    
    content.innerHTML = `
        <div class="flex items-center justify-center py-8">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            <span class="ml-2 text-gray-600">Ładowanie szczegółów...</span>
        </div>
    `;
    
    fetch(`/api/notifications/details/${groupId}?tab=${currentNotificationTab}`)
        .then(response => response.text())
        .then(html => {
            content.innerHTML = html;
        })
        .catch(error => {
            console.error('Error loading notification group:', error);
            content.innerHTML = `
                <div class="flex items-center justify-center py-8 text-red-500">
                    <p>Błąd ładowania szczegółów</p>
                </div>
            `;
        });
}

// Mark group as read
function markGroupAsRead(groupId, event) {
    if (event) {
        event.stopPropagation();
    }
    
    fetch(`/api/notifications/groups/${groupId}/read`, {
        method: 'POST'
    })
    .then(response => {
        if (response.ok) {
            // Reload notification bell
            loadNotificationBell();
            
            // If we're viewing the group, reload it
            if (currentNotificationGroup === groupId) {
                openNotificationGroup(groupId);
            } else {
                // Reload the groups list
                loadNotificationGroups(currentNotificationTab);
            }
        }
    })
    .catch(error => {
        console.error('Error marking group as read:', error);
    });
}

// Helper functions - no longer needed as HTML is generated server-side

// Auto-refresh notifications every 30 seconds
setInterval(() => {
    if (document.getElementById('notification-sidebar') && 
        !document.getElementById('notification-sidebar').classList.contains('translate-x-full')) {
        loadNotificationBell();
        if (currentNotificationGroup) {
            openNotificationGroup(currentNotificationGroup);
        } else {
            loadNotificationGroups(currentNotificationTab);
        }
    } else {
        loadNotificationBell();
    }
}, 30000);

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    initializeNotifications();
});

// Make functions globally available
window.openNotificationSidebar = openNotificationSidebar;
window.closeNotificationSidebar = closeNotificationSidebar;
window.switchNotificationTab = switchNotificationTab;
window.openNotificationGroup = openNotificationGroup;
window.markGroupAsRead = markGroupAsRead; 