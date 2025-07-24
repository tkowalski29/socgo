// Global sidebar functions
function openPostSidebar() {
  const sidebar = document.getElementById('post-create-sidebar');
  if (!sidebar) {
    console.error('Sidebar element not found!');
    return;
  }

  console.log('Adding open class to sidebar');
  console.log('Current classes before:', sidebar.classList.toString());
  sidebar.classList.add('open');
  console.log('Current classes after:', sidebar.classList.toString());
  console.log('Computed transform:', getComputedStyle(sidebar).transform);
  
  // Force the transform to ensure sidebar is visible
  sidebar.style.transform = 'translateY(0)';
  console.log('Forced transform to translateY(0)');
  
  // Load providers when sidebar opens
  if (typeof loadProviders === 'function') {
    loadProviders();
  }
  
  // Focus on content textarea
  setTimeout(() => {
    const textarea = document.getElementById('content');
    if (textarea) {
      textarea.focus();
    }
  }, 300);
}

// Function to open sidebar with specific date and time
function openPostSidebarWithDateTime(dateStr, hour) {
  console.log('openPostSidebarWithDateTime called with:', dateStr, hour);
  
  // Check if DOM is ready
  if (document.readyState === 'loading') {
    console.log('DOM not ready, waiting...');
    document.addEventListener('DOMContentLoaded', () => {
      openPostSidebarWithDateTime(dateStr, hour);
    });
    return;
  }
  
  // Check if sidebar element exists
  const sidebar = document.getElementById('post-create-sidebar');
  if (!sidebar) {
    console.error('Sidebar element not found!');
    return;
  }
  
  // Check if openPostSidebar function exists
  if (typeof openPostSidebar === 'function') {
    console.log('openPostSidebar function found, calling it...');
    openPostSidebar();
  } else {
    console.error('openPostSidebar function not found!');
    // Fallback: manually open sidebar
    console.log('Manually opening sidebar...');
    sidebar.classList.add('open');
    sidebar.style.transform = 'translateY(0)';
  }
  
  // Set the datetime to the specified date and hour
  setTimeout(() => {
    const date = new Date(dateStr);
    date.setHours(hour, 0, 0, 0); // Set to the specified hour, minutes to 0
    
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    
    const displayDateTime = `${year}-${month}-${day} ${hours}:${minutes}`;
    const nativeDateTime = `${year}-${month}-${day}T${hours}:${minutes}`;
    
    // Set display input
    const footerScheduleInput = document.getElementById('schedule_at_footer');
    if (footerScheduleInput) footerScheduleInput.value = displayDateTime;
    
    // Set native input for form submission
    const nativeInput = document.getElementById('schedule_at_native');
    if (nativeInput) nativeInput.value = nativeDateTime;
    
    console.log('DateTime set to:', displayDateTime);
  }, 100);
}

function closePostSidebar() {
  const sidebar = document.getElementById('post-create-sidebar');
  if (!sidebar) {
    return;
  }
  
  sidebar.classList.remove('open');
  
  // Clear form when closing
  const form = document.getElementById('post-form');
  if (form) {
    form.reset();
    document.getElementById('file-preview').innerHTML = '';
    document.getElementById('provider-settings-section').classList.add('hidden');
    document.getElementById('post-result').innerHTML = '';
  }
}

// Set default datetime to current date and time
function setDefaultDateTime() {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, '0');
  const day = String(now.getDate()).padStart(2, '0');
  const hours = String(now.getHours()).padStart(2, '0');
  const minutes = String(now.getMinutes()).padStart(2, '0');
  
  const defaultDateTime = `${year}-${month}-${day}T${hours}:${minutes}`;
  const displayDateTime = `${year}-${month}-${day} ${hours}:${minutes}`;
  
  // Set display input
  const footerScheduleInput = document.getElementById('schedule_at_footer');
  if (footerScheduleInput) footerScheduleInput.value = displayDateTime;
  
  // Set native input for form submission
  const nativeInput = document.getElementById('schedule_at_native');
  if (nativeInput) nativeInput.value = defaultDateTime;
}

// Format datetime for display
function formatDateTimeForDisplay(date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  
  return `${year}-${month}-${day} ${hours}:${minutes}`;
}

// Format datetime for native input
function formatDateTimeForNative(date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  
  return `${year}-${month}-${day}T${hours}:${minutes}`;
}

// Handle file selection and preview
function handleFileSelect(input) {
  const preview = document.getElementById('file-preview');
  preview.innerHTML = '';
  
  if (input.files) {
    Array.from(input.files).forEach((file, index) => {
      const reader = new FileReader();
      const previewItem = document.createElement('div');
      previewItem.className = 'relative group';
      
      reader.onload = function(e) {
        const isVideo = file.type.startsWith('video/');
        const isImage = file.type.startsWith('image/');
        
        if (isImage) {
          previewItem.innerHTML = `
            <div class="relative w-16 h-16">
              <img src="${e.target.result}" alt="${file.name}" class="w-full h-full object-cover rounded-lg">
              <button type="button" onclick="removeFile(${index})" class="absolute -top-1 -right-1 bg-red-500 text-white rounded-full w-5 h-5 flex items-center justify-center text-xs hover:bg-red-600">
                ×
              </button>
            </div>
          `;
        } else if (isVideo) {
          previewItem.innerHTML = `
            <div class="relative w-16 h-16">
              <video src="${e.target.result}" class="w-full h-full object-cover rounded-lg"></video>
              <button type="button" onclick="removeFile(${index})" class="absolute -top-1 -right-1 bg-red-500 text-white rounded-full w-5 h-5 flex items-center justify-center text-xs hover:bg-red-600">
                ×
              </button>
            </div>
          `;
        }
      };
      
      reader.readAsDataURL(file);
      preview.appendChild(previewItem);
    });
  }
}

// Remove file from preview
function removeFile(index) {
  const input = document.getElementById('media');
  const dt = new DataTransfer();
  
  Array.from(input.files).forEach((file, i) => {
    if (i !== index) {
      dt.items.add(file);
    }
  });
  
  input.files = dt.files;
  handleFileSelect(input);
}



// Handle provider selection and load settings
function handleProviderSelection() {
  const selectedProviders = document.querySelectorAll('.provider-checkbox:checked');
  const settingsSection = document.getElementById('provider-settings-section');
  const settingsContainer = document.getElementById('provider-settings-container');
  const tabsContainer = document.getElementById('provider-tabs');
  
  // Update visual state of provider icons
  document.querySelectorAll('.provider-checkbox').forEach(checkbox => {
    const label = checkbox.closest('label');
    const iconDiv = label.querySelector('div');
    if (checkbox.checked) {
      iconDiv.classList.add('ring-2', 'ring-blue-500', 'ring-offset-2');
      iconDiv.classList.remove('border-gray-200');
      iconDiv.classList.add('border-blue-500');
    } else {
      iconDiv.classList.remove('ring-2', 'ring-blue-500', 'ring-offset-2');
      iconDiv.classList.remove('border-blue-500');
      iconDiv.classList.add('border-gray-200');
    }
  });
  
  if (selectedProviders.length > 0) {
    settingsSection.classList.remove('hidden');
    settingsContainer.innerHTML = '';
    tabsContainer.innerHTML = '';
    
    // Create tabs
    selectedProviders.forEach((checkbox, index) => {
      const providerType = checkbox.getAttribute('data-provider-type');
      const providerName = checkbox.getAttribute('data-provider-name');
      const providerIcon = checkbox.getAttribute('data-provider-icon');
      
      const tab = document.createElement('button');
      tab.className = `px-3 py-2 text-sm font-medium rounded-md transition-colors ${index === 0 ? 'bg-blue-100 text-blue-700' : 'text-gray-500 hover:text-gray-700'}`;
      tab.textContent = providerName;
      tab.onclick = () => switchProviderTab(index, providerType, providerName);
      tabsContainer.appendChild(tab);
      
      // Load first provider settings by default
      if (index === 0) {
        loadProviderSettings(providerType, providerName, settingsContainer);
      }
    });
  } else {
    settingsSection.classList.add('hidden');
  }
}

function switchProviderTab(index, providerType, providerName) {
  const tabs = document.querySelectorAll('#provider-tabs button');
  const settingsContainer = document.getElementById('provider-settings-container');
  
  // Update tab styles
  tabs.forEach((tab, i) => {
    if (i === index) {
      tab.className = 'px-3 py-2 text-sm font-medium rounded-md transition-colors bg-blue-100 text-blue-700';
    } else {
      tab.className = 'px-3 py-2 text-sm font-medium rounded-md transition-colors text-gray-500 hover:text-gray-700';
    }
  });
  
  // Load provider settings
  loadProviderSettings(providerType, providerName, settingsContainer);
}

function loadProviderSettings(providerType, providerName, container) {
  container.innerHTML = '<div class="flex items-center justify-center py-4"><div class="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div><span class="ml-2 text-sm text-gray-600">Ładowanie ustawień...</span></div>';
  
  fetch(`/api/providers/settings?type=${providerType}&name=${encodeURIComponent(providerName)}`)
    .then(response => response.text())
    .then(html => {
      container.innerHTML = html;
    })
    .catch(error => {
      console.error('Error loading provider settings:', error);
      container.innerHTML = '<p class="text-red-600 text-sm">Błąd ładowania ustawień</p>';
    });
}

// Load providers on sidebar open
function loadProviders() {
  fetch('/api/providers/options')
    .then(response => response.text())
    .then(html => {
      const container = document.getElementById('providers-container');
      container.innerHTML = html;
      
      // Add event listeners to provider checkboxes and labels
      document.querySelectorAll('.provider-checkbox').forEach(checkbox => {
        checkbox.addEventListener('change', handleProviderSelection);
      });
      
      // Add click handlers for labels to toggle checkboxes
      document.querySelectorAll('label[class*="cursor-pointer"]').forEach(label => {
        label.addEventListener('click', function(e) {
          const checkbox = this.querySelector('input[type="checkbox"]');
          if (checkbox && !checkbox.disabled) {
            checkbox.checked = !checkbox.checked;
            checkbox.dispatchEvent(new Event('change'));
          }
        });
      });
    })
    .catch(error => {
      console.error('Error loading providers:', error);
      document.getElementById('providers-container').innerHTML = 
        '<p class="text-red-600 text-sm">Błąd ładowania platform</p>';
    });
}

// Function to add another post
function addAnotherPost() {
  const form = document.getElementById('post-form');
  const resultDiv = document.getElementById('post-result');
  
  // Clear form and result
  form.reset();
  document.getElementById('file-preview').innerHTML = '';
  document.getElementById('provider-settings-section').classList.add('hidden');
  resultDiv.innerHTML = '';
  
  // Reset datetime to current
  setDefaultDateTime();
  
  // Focus on content textarea
  setTimeout(() => {
    const textarea = document.getElementById('content');
    if (textarea) {
      textarea.focus();
    }
  }, 100);
}

// Initialize datetime on page load
document.addEventListener('DOMContentLoaded', function() {
  setDefaultDateTime();
  
  // Custom datetime picker functionality
  const datetimeInput = document.getElementById('schedule_at_footer');
  const pickerBtn = document.getElementById('datetime-picker-btn');
  const customPicker = document.getElementById('custom-datetime-picker');
  const closeBtn = document.getElementById('close-datetime-picker');
  const cancelBtn = document.getElementById('cancel-datetime');
  const applyBtn = document.getElementById('apply-datetime');
  const dateInput = document.getElementById('custom-date');
  const hourInput = document.getElementById('custom-hour');
  const minuteInput = document.getElementById('custom-minute');
  
  // Open custom picker
  function openCustomPicker() {
    const currentValue = datetimeInput.value;
    if (currentValue) {
      const [datePart, timePart] = currentValue.split(' ');
      if (datePart && timePart) {
        const [year, month, day] = datePart.split('-');
        const [hour, minute] = timePart.split(':');
        
        dateInput.value = `${year}-${month}-${day}`;
        hourInput.value = hour;
        minuteInput.value = minute;
      }
    } else {
      const now = new Date();
      dateInput.value = now.toISOString().split('T')[0];
      hourInput.value = String(now.getHours()).padStart(2, '0');
      minuteInput.value = String(now.getMinutes()).padStart(2, '0');
    }
    
    customPicker.classList.add('show');
  }
  
  // Close custom picker
  function closeCustomPicker() {
    customPicker.classList.remove('show');
  }
  
  // Apply datetime
  function applyDateTime() {
    const date = dateInput.value;
    const hour = hourInput.value.padStart(2, '0');
    const minute = minuteInput.value.padStart(2, '0');
    
    if (date && hour && minute) {
      const displayValue = `${date} ${hour}:${minute}`;
      const nativeValue = `${date}T${hour}:${minute}`;
      
      datetimeInput.value = displayValue;
      document.getElementById('schedule_at_native').value = nativeValue;
    }
    
    closeCustomPicker();
  }
  
  // Event listeners
  if (datetimeInput) {
    datetimeInput.addEventListener('click', openCustomPicker);
  }
  
  if (pickerBtn) {
    pickerBtn.addEventListener('click', openCustomPicker);
  }
  
  if (closeBtn) {
    closeBtn.addEventListener('click', function(e) {
      e.preventDefault();
      e.stopPropagation();
      closeCustomPicker();
    });
  }
  
  if (cancelBtn) {
    cancelBtn.addEventListener('click', function(e) {
      e.preventDefault();
      e.stopPropagation();
      closeCustomPicker();
    });
  }
  
  if (applyBtn) {
    applyBtn.addEventListener('click', applyDateTime);
  }
  
  // Close picker when clicking outside
  document.addEventListener('click', function(e) {
    if (!customPicker.contains(e.target) && 
        !datetimeInput.contains(e.target) && 
        !pickerBtn.contains(e.target)) {
      closeCustomPicker();
    }
  });
  
  // Handle Enter key in inputs
  [dateInput, hourInput, minuteInput].forEach(input => {
    if (input) {
      input.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
          applyDateTime();
        }
      });
    }
  });
});

// Form submission handling
document.addEventListener('DOMContentLoaded', function() {
  const form = document.getElementById('post-form');
  const resultDiv = document.getElementById('post-result');
  const submitBtn = document.getElementById('submit-btn');
  
  form.addEventListener('htmx:beforeRequest', function() {
    submitBtn.disabled = true;
    submitBtn.innerHTML = '<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>Wysyłanie...';
    resultDiv.innerHTML = '';
  });
  
  form.addEventListener('htmx:afterRequest', function(event) {
    submitBtn.disabled = false;
    submitBtn.innerHTML = '<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"></path></svg>Opublikuj';
    
    if (event.detail.successful) {
      resultDiv.innerHTML = `
        <div class="text-green-600 text-sm mb-2">✓ Post został wysłany pomyślnie!</div>
        <div class="flex space-x-2">
          <button
            type="button"
            onclick="closePostSidebar()"
            class="px-3 py-1 text-xs bg-gray-100 text-gray-700 rounded hover:bg-gray-200 transition-colors"
          >
            Zamknij
          </button>
          <button
            type="button"
            onclick="addAnotherPost()"
            class="px-3 py-1 text-xs bg-blue-100 text-blue-700 rounded hover:bg-blue-200 transition-colors"
          >
            Dodaj kolejny
          </button>
        </div>
      `;
      // Clear form
      form.reset();
      document.getElementById('file-preview').innerHTML = '';
      document.getElementById('provider-settings-section').classList.add('hidden');
      setDefaultDateTime();
    } else {
      resultDiv.innerHTML = '<div class="text-red-600 text-sm">✗ Błąd podczas wysyłania posta</div>';
    }
  });
});

// Make functions globally available immediately
window.openPostSidebar = openPostSidebar;
window.openPostSidebarWithDateTime = openPostSidebarWithDateTime;
window.closePostSidebar = closePostSidebar;

// Make all functions globally available
window.openPostSidebar = openPostSidebar;
window.closePostSidebar = closePostSidebar;
window.handleFileSelect = handleFileSelect;
window.removeFile = removeFile;
window.handleProviderSelection = handleProviderSelection;
window.switchProviderTab = switchProviderTab;
window.loadProviders = loadProviders;
window.addAnotherPost = addAnotherPost;