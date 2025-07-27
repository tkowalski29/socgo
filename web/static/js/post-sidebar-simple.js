// Simple Post Sidebar Handler - VERSION 20250727
document.addEventListener('DOMContentLoaded', function() {
  console.log('Post sidebar simple handler loaded - VERSION 20250727');
  
  // Initialize when sidebar opens
  window.openPostSidebar = function() {
    console.log('openPostSidebar called');
    const sidebar = document.getElementById('post-create-sidebar');
    console.log('Sidebar element:', sidebar);
    
    if (sidebar) {
      console.log('Opening sidebar');
      sidebar.classList.add('open');
      sidebar.style.transform = 'translateY(0)';
      
      // Load providers
      console.log('Loading providers...');
      loadProviders();
      
      // Initialize preview
      console.log('Initializing preview...');
      initializePreview();
      
      // Focus on textarea
      setTimeout(() => {
        const textarea = document.getElementById('content');
        if (textarea) {
          console.log('Focusing on textarea');
          textarea.focus();
        } else {
          console.log('Textarea not found');
        }
      }, 300);
    } else {
      console.error('Sidebar element not found!');
    }
  };
  
  // Close sidebar
  window.closePostSidebar = function() {
    const sidebar = document.getElementById('post-create-sidebar');
    if (sidebar) {
      sidebar.classList.remove('open');
    }
  };
  
  // Load providers
  function loadProviders() {
    console.log('loadProviders called');
    const container = document.getElementById('providers-container');
    console.log('Providers container:', container);
    
    if (!container) {
      console.error('Providers container not found!');
      return;
    }
    
    console.log('Fetching providers from /api/providers/options');
    fetch('/api/providers/options')
      .then(response => {
        console.log('Providers response status:', response.status);
        return response.text();
      })
      .then(html => {
        console.log('Providers HTML received, length:', html.length);
        console.log('Providers HTML preview:', html.substring(0, 200));
        container.innerHTML = html;
        setupProviderHandlers();
      })
      .catch(error => {
        console.error('Error loading providers:', error);
        container.innerHTML = '<p class="text-red-600 text-sm">Błąd ładowania platform</p>';
      });
  }
  
  // Setup provider click handlers
  function setupProviderHandlers() {
    const labels = document.querySelectorAll('#providers-container label');
    console.log('Setting up provider handlers, found labels:', labels.length);
    
    labels.forEach((label, index) => {
      console.log(`Setting up handler for label ${index}:`, label);
      
      const checkbox = label.querySelector('input[type="checkbox"]');
      console.log(`Checkbox ${index} disabled:`, checkbox.disabled);
      console.log(`Checkbox ${index} attributes:`, checkbox.attributes);
      
      label.addEventListener('click', function(e) {
        console.log('Label clicked:', this);
        e.preventDefault();
        
        const checkbox = this.querySelector('input[type="checkbox"]');
        console.log('Found checkbox:', checkbox);
        console.log('Checkbox disabled:', checkbox.disabled);
        
        if (checkbox && !checkbox.disabled) {
          console.log('Checkbox before toggle:', checkbox.checked);
          // Toggle checkbox
          checkbox.checked = !checkbox.checked;
          console.log('Checkbox after toggle:', checkbox.checked);
          
          // Update visual state
          updateProviderVisualState();
          
          // Handle provider selection
          handleProviderSelection();
        } else {
          console.log('Checkbox not found or disabled');
        }
      });
    });
  }
  
  // Update visual state of provider icons
  function updateProviderVisualState() {
    console.log('updateProviderVisualState called');
    const checkboxes = document.querySelectorAll('.provider-checkbox');
    console.log('Found checkboxes:', checkboxes.length);
    
    checkboxes.forEach((checkbox, index) => {
      const label = checkbox.closest('label');
      const iconDiv = label.querySelector('div');
      
      console.log(`Checkbox ${index}:`, checkbox.checked);
      
      if (checkbox.checked) {
        iconDiv.classList.add('ring-2', 'ring-blue-500', 'ring-offset-2', 'border-blue-500');
        iconDiv.classList.remove('border-gray-200');
      } else {
        iconDiv.classList.remove('ring-2', 'ring-blue-500', 'ring-offset-2', 'border-blue-500');
        iconDiv.classList.add('border-gray-200');
      }
    });
  }
  
  // Handle provider selection
  // Clear provider settings cache
  function clearProviderSettingsCache() {
    providerSettingsCache = {};
    console.log('Cleared provider settings cache');
  }
  
  function handleProviderSelection() {
    console.log('handleProviderSelection called');
    
    const selectedProviders = document.querySelectorAll('.provider-checkbox:checked');
    const tabsSection = document.getElementById('tabs-section');
    
    console.log('Selected providers:', selectedProviders.length);
    console.log('Tabs section:', tabsSection);
    
    if (selectedProviders.length > 0) {
      console.log('Showing provider settings with preview tab');
      // Show tabs section
      tabsSection.classList.remove('hidden');
      
      // Save current provider settings before changing
      saveCurrentProviderSettings();
      
      // Check if this is a new provider being added
      const wasEmpty = document.querySelectorAll('#provider-tabs button').length <= 1; // Only preview tab
      const isNewProvider = wasEmpty && selectedProviders.length === 1;
      
      if (isNewProvider) {
        console.log('New provider added, will switch to its tab');
        // Don't clear cache for new provider
        loadProviderSettingsWithActiveTab(selectedProviders, 1); // Switch to provider tab (index 1)
      } else {
        // Clear cache when provider selection changes (removing providers)
        clearProviderSettingsCache();
        loadProviderSettings(selectedProviders);
      }
    } else {
      console.log('Showing only preview');
      // Show tabs section with only preview tab
      tabsSection.classList.remove('hidden');
      showOnlyPreviewTab();
    }
  }
  
  // Load provider settings with specific active tab
  function loadProviderSettingsWithActiveTab(selectedProviders, activeTabIndex) {
    console.log('loadProviderSettingsWithActiveTab called with:', selectedProviders.length, 'providers, active tab:', activeTabIndex);
    
    const tabsContainer = document.getElementById('provider-tabs');
    
    console.log('Tabs container:', tabsContainer);
    
    // Clear containers
    tabsContainer.innerHTML = '';
    
    // Get provider IDs
    const providerIDs = Array.from(selectedProviders).map(cb => cb.value).join(',');
    console.log('Provider IDs:', providerIDs);
    
    // Load tabs with specific active tab
    const tabsUrl = `/api/providers/tabs?providers=${providerIDs}&active=${activeTabIndex}`;
    console.log('Loading tabs from:', tabsUrl);
    
    fetch(tabsUrl)
      .then(response => {
        console.log('Tabs response status:', response.status);
        return response.text();
      })
      .then(html => {
        console.log('Tabs HTML received, length:', html.length);
        
        // Add preview tab at the beginning (icon only)
        const previewTabButton = `
            <button class="flex items-center justify-center w-10 h-10 rounded-md transition-colors text-gray-500 hover:text-gray-700" data-index="preview" onclick="switchToPreviewTab()" title="Preview">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
              </svg>
            </button>
        `;
        
        tabsContainer.innerHTML = previewTabButton + html;
        
        // Switch to the specified active tab
        if (activeTabIndex === 0) {
          showPreviewTab();
        } else {
          // Find the provider tab and switch to it
          const providerTab = document.querySelector(`#provider-tabs button[data-index="${activeTabIndex}"]`);
          if (providerTab) {
            const providerType = providerTab.dataset.type;
            const providerName = providerTab.dataset.name;
            switchProviderTab(activeTabIndex, providerType, providerName);
          }
        }
      })
      .catch(error => {
        console.error('Error loading tabs:', error);
        tabsContainer.innerHTML = '<p class="text-red-600 text-sm">Błąd ładowania zakładek</p>';
      });
  }
  
  // Load provider settings
  function loadProviderSettings(selectedProviders) {
    console.log('loadProviderSettings called with:', selectedProviders.length, 'providers');
    
    const tabsContainer = document.getElementById('provider-tabs');
    
    console.log('Tabs container:', tabsContainer);
    
    // Clear containers
    tabsContainer.innerHTML = '';
    
    // Get provider IDs
    const providerIDs = Array.from(selectedProviders).map(cb => cb.value).join(',');
    console.log('Provider IDs:', providerIDs);
    
    // Load tabs
    const tabsUrl = `/api/providers/tabs?providers=${providerIDs}&active=0`;
    console.log('Loading tabs from:', tabsUrl);
    
    fetch(tabsUrl)
      .then(response => {
        console.log('Tabs response status:', response.status);
        return response.text();
      })
      .then(html => {
        console.log('Tabs HTML received, length:', html.length);
        
        // Add preview tab at the beginning (icon only)
        const previewTabButton = `
            <button class="flex items-center justify-center w-10 h-10 rounded-md transition-colors bg-blue-100 text-blue-700" data-index="preview" onclick="switchToPreviewTab()" title="Preview">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
              </svg>
            </button>
        `;
        
        tabsContainer.innerHTML = previewTabButton + html;
        
        // Show preview by default
        showPreviewTab();
      })
      .catch(error => {
        console.error('Error loading tabs:', error);
        tabsContainer.innerHTML = '<p class="text-red-700 text-sm">Błąd ładowania zakładek</p>';
      });
  }
  
  // Load single provider settings
  // Save current provider settings to cache
  function saveCurrentProviderSettings() {
    console.log('saveCurrentProviderSettings called');
    const tabContent = document.getElementById('tab-content');
    if (!tabContent) {
      console.log('No tab content found');
      return;
    }
    
    // Find the currently active tab
    const activeTab = document.querySelector('#provider-tabs button[class*="bg-blue-100"]');
    console.log('Active tab found:', activeTab);
    if (!activeTab) {
      console.log('No active tab found');
      return;
    }
    
    const providerType = activeTab.dataset.type;
    const providerName = activeTab.dataset.name;
    console.log('Provider type:', providerType, 'Provider name:', providerName);
    
    if (!providerType || !providerName) {
      console.log('Missing provider type or name');
      return;
    }
    
    const cacheKey = `${providerType}_${providerName}`;
    const settings = {};
    
    // Save all input values in the current tab content
    const inputs = tabContent.querySelectorAll('input, select, textarea');
    console.log('Found inputs:', inputs.length);
    inputs.forEach((input, index) => {
      if (input.name) {
        if (input.type === 'checkbox') {
          settings[input.name] = input.checked;
          console.log(`Input ${index}: ${input.name} (checkbox) = ${input.checked}`);
        } else {
          settings[input.name] = input.value;
          console.log(`Input ${index}: ${input.name} = "${input.value}"`);
        }
      }
    });
    
    providerSettingsCache[cacheKey] = settings;
    console.log('Saved settings for', cacheKey, ':', settings);
    console.log('Current cache:', providerSettingsCache);
  }
  
  // Restore provider settings from cache
  function restoreProviderSettings(providerType, providerName) {
    console.log('restoreProviderSettings called for:', providerType, providerName);
    const cacheKey = `${providerType}_${providerName}`;
    const settings = providerSettingsCache[cacheKey];
    
    console.log('Cache key:', cacheKey);
    console.log('Settings found:', settings);
    
    if (!settings) {
      console.log('No settings found in cache');
      return;
    }
    
    console.log('Restoring settings for', cacheKey, ':', settings);
    
    // Wait a bit for the DOM to be updated
    setTimeout(() => {
      const tabContent = document.getElementById('tab-content');
      if (!tabContent) {
        console.log('No tab content found during restore');
        return;
      }
      
      // Restore all input values
      const inputs = tabContent.querySelectorAll('input, select, textarea');
      console.log('Found inputs during restore:', inputs.length);
      inputs.forEach((input, index) => {
        if (input.name && settings.hasOwnProperty(input.name)) {
          if (input.type === 'checkbox') {
            input.checked = settings[input.name];
            console.log(`Restored input ${index}: ${input.name} (checkbox) = ${settings[input.name]}`);
          } else {
            input.value = settings[input.name];
            console.log(`Restored input ${index}: ${input.name} = "${settings[input.name]}"`);
          }
        }
      });
    }, 100);
  }
  
  function loadSingleProviderSettings(providerType, providerName, container) {
    console.log('loadSingleProviderSettings called for:', providerType, providerName);
    console.log('Container:', container);
    
    const tabContent = document.getElementById('tab-content');
    
    tabContent.innerHTML = '<div class="flex items-center justify-center py-4"><div class="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div><span class="ml-2 text-sm text-gray-600">Ładowanie ustawień...</span></div>';
    
    const settingsUrl = `/api/providers/settings?type=${providerType}&name=${encodeURIComponent(providerName)}`;
    console.log('Loading settings from:', settingsUrl);
    
    fetch(settingsUrl)
      .then(response => {
        console.log('Settings response status:', response.status);
        return response.text();
      })
      .then(html => {
        console.log('Settings HTML received, length:', html.length);
        console.log('Settings HTML preview:', html.substring(0, 200));
        tabContent.innerHTML = html;
        
        // Restore cached settings after loading
        restoreProviderSettings(providerType, providerName);
      })
      .catch(error => {
        console.error('Error loading settings:', error);
        tabContent.innerHTML = '<p class="text-red-600 text-sm">Błąd ładowania ustawień</p>';
      });
  }
  
  // Initialize preview functionality
  function initializePreview() {
    console.log('initializePreview called');
    const textarea = document.getElementById('content');
    const fileInput = document.getElementById('media');
    const tabsSection = document.getElementById('tabs-section');
    
    console.log('Textarea found:', !!textarea);
    console.log('File input found:', !!fileInput);
    console.log('Tabs section found:', !!tabsSection);
    
    // Show tabs section with preview tab initially
    if (tabsSection) {
      tabsSection.classList.remove('hidden');
      // Show preview tab initially
      showOnlyPreviewTab();
    }
    
    if (textarea) {
      textarea.addEventListener('input', updatePreview);
      console.log('Added input listener to textarea');
    }
    
    if (fileInput) {
      fileInput.addEventListener('change', updatePreview);
      console.log('Added change listener to file input');
    }
    
    // Initial preview update
    console.log('Updating initial preview');
    updatePreview();
  }
  
  // Update preview
  function updatePreview() {
    console.log('updatePreview called');
    updateTextPreview();
    updateMediaPreview();
  }
  
  // Update text preview
  function updateTextPreview() {
    console.log('updateTextPreview called');
    const textarea = document.getElementById('content');
    const previewContainer = document.getElementById('post-content-preview');
    
    console.log('Textarea found:', !!textarea);
    console.log('Preview container found:', !!previewContainer);
    
    if (!textarea || !previewContainer) return;
    
    const content = textarea.value.trim();
    console.log('Content:', content);
    
    if (content) {
      previewContainer.innerHTML = `<p class="text-gray-900 whitespace-pre-wrap">${content}</p>`;
    } else {
      previewContainer.innerHTML = '<p class="text-gray-500 text-sm">Treść posta pojawi się tutaj...</p>';
    }
  }
  
  // Update media preview
  function updateMediaPreview() {
    console.log('updateMediaPreview called');
    const fileInput = document.getElementById('media');
    const mediaGrid = document.getElementById('post-media-grid');
    
    console.log('File input found:', !!fileInput);
    console.log('Media grid found:', !!mediaGrid);
    
    if (!fileInput || !mediaGrid) return;
    
    mediaGrid.innerHTML = '';
    
    if (fileInput.files.length === 0) {
      console.log('No files selected');
      return;
    }
    
    console.log('Files selected:', fileInput.files.length);
    Array.from(fileInput.files).forEach(file => {
      console.log('Processing file:', file.name, file.type);
      if (file.type.startsWith('image/')) {
        const reader = new FileReader();
        reader.onload = function(e) {
          const img = document.createElement('img');
          img.src = e.target.result;
          img.className = 'w-full h-24 object-cover rounded-lg';
          img.alt = file.name;
          mediaGrid.appendChild(img);
        };
        reader.readAsDataURL(file);
      } else if (file.type.startsWith('video/')) {
        const video = document.createElement('video');
        video.src = URL.createObjectURL(file);
        video.className = 'w-full h-24 object-cover rounded-lg';
        video.controls = true;
        video.muted = true;
        mediaGrid.appendChild(video);
      }
    });
  }
  
  // Store provider settings values
  let providerSettingsCache = {};
  
  // Global function for switching provider tabs
  window.switchProviderTab = function(index, providerType, providerName) {
    console.log('switchProviderTab called:', index, providerType, providerName);
    
    // Save current provider settings before switching
    console.log('About to save current settings...');
    saveCurrentProviderSettings();
    
    // Load provider settings
    console.log('About to load new settings...');
    loadSingleProviderSettings(providerType, providerName);
    
    // Update tab styles without reloading all tabs
    updateTabStyles(index.toString());
  };
  
  // Submit form
  window.submitPostForm = function() {
    console.log('submitPostForm called');
    const form = document.getElementById('post-form');
    console.log('Form element:', form);
    
    if (form) {
      // Add provider settings to form
      addProviderSettingsToForm(form);
      
      // Show loading state
      const submitBtn = document.getElementById('submit-btn');
      const submitBtnLoading = document.getElementById('submit-btn-loading');
      
      console.log('Submit button:', submitBtn);
      console.log('Submit button loading:', submitBtnLoading);
      
      if (submitBtn) submitBtn.classList.add('hidden');
      if (submitBtnLoading) submitBtnLoading.classList.remove('hidden');
      
      // Clear result div
      const resultDiv = document.getElementById('post-result');
      if (resultDiv) resultDiv.innerHTML = '';
      
      // Submit form
      console.log('Submitting form');
      form.requestSubmit();
    } else {
      console.error('Form not found!');
    }
  };
  
  // Handle HTMX events
  document.addEventListener('htmx:beforeRequest', function(event) {
    console.log('HTMX before request:', event);
  });
  
  document.addEventListener('htmx:afterRequest', function(event) {
    console.log('HTMX after request:', event);
    
    // Show normal button, hide loading button
    const submitBtn = document.getElementById('submit-btn');
    const submitBtnLoading = document.getElementById('submit-btn-loading');
    
    if (submitBtn) submitBtn.classList.remove('hidden');
    if (submitBtnLoading) submitBtnLoading.classList.add('hidden');
    
    // Check if the request was successful (HTTP 200-299)
    const isSuccessful = event.detail.xhr.status >= 200 && event.detail.xhr.status < 300;
    
    if (isSuccessful) {
      // Hide the footer
      const footerContainer = document.querySelector('.p-4.border-t.border-gray-200.bg-gray-50');
      if (footerContainer) {
        footerContainer.style.display = 'none';
      }
    }
  });
  
  // Add provider settings to form
  function addProviderSettingsToForm(form) {
    console.log('addProviderSettingsToForm called');
    
    // Add selected providers
    const selectedProviders = document.querySelectorAll('.provider-checkbox:checked');
    const providerIDs = Array.from(selectedProviders).map(cb => cb.value);
    console.log('Selected provider IDs:', providerIDs);
    
    // Remove existing providers input
    const existingProvidersInput = form.querySelector('input[name="providers"]');
    if (existingProvidersInput) {
      existingProvidersInput.remove();
    }
    
    // Create providers input
    if (providerIDs.length > 0) {
      const providersInput = document.createElement('input');
      providersInput.type = 'hidden';
      providersInput.name = 'providers';
      providersInput.value = providerIDs.join(',');
      form.appendChild(providersInput);
    }
    
    // Add schedule_at_native field
    const scheduleInput = document.getElementById('schedule_at_native');
    if (scheduleInput && scheduleInput.value) {
      // Remove existing schedule input
      const existingScheduleInput = form.querySelector('input[name="schedule_at_native"]');
      if (existingScheduleInput) {
        existingScheduleInput.remove();
      }
      
      // Create schedule input
      const scheduleHiddenInput = document.createElement('input');
      scheduleHiddenInput.type = 'hidden';
      scheduleHiddenInput.name = 'schedule_at_native';
      scheduleHiddenInput.value = scheduleInput.value;
      form.appendChild(scheduleHiddenInput);
    }
    
    // Add provider settings
    const settingsInputs = document.querySelectorAll('#tab-content input, #tab-content select, #tab-content textarea');
    console.log('Found settings inputs:', settingsInputs.length);
    
    settingsInputs.forEach((input, index) => {
      console.log(`Input ${index}:`, input.name, input.value);
      // Remove existing hidden inputs with same name
      const existingInputs = form.querySelectorAll(`input[name="${input.name}"]`);
      existingInputs.forEach(existing => {
        if (existing.type === 'hidden') {
          existing.remove();
        }
      });
      
      // Create hidden input
      const hiddenInput = document.createElement('input');
      hiddenInput.type = 'hidden';
      hiddenInput.name = input.name;
      hiddenInput.value = input.value;
      form.appendChild(hiddenInput);
    });
  }
  
  // Handle file selection
  window.handleFileSelect = function(input) {
    console.log('handleFileSelect called');
    const preview = document.getElementById('file-preview');
    console.log('File preview element:', preview);
    
    if (!preview) return;
    
    preview.innerHTML = '';
    
    if (input.files) {
      console.log('Files selected:', input.files.length);
      Array.from(input.files).forEach((file, index) => {
        console.log(`File ${index}:`, file.name, file.type);
        const reader = new FileReader();
        reader.onload = function(e) {
          const previewItem = document.createElement('div');
          previewItem.className = 'relative group';
          
          if (file.type.startsWith('image/')) {
            previewItem.innerHTML = `
              <div class="relative w-16 h-16">
                <img src="${e.target.result}" alt="${file.name}" class="w-full h-full object-cover rounded-lg">
                <button type="button" data-index="${index}" onclick="removeFile(${index})" class="absolute -top-1 -right-1 bg-red-500 text-white rounded-full w-5 h-5 flex items-center justify-center text-xs hover:bg-red-600">
                  ×
                </button>
              </div>
            `;
          } else if (file.type.startsWith('video/')) {
            previewItem.innerHTML = `
              <div class="relative w-16 h-16">
                <video src="${e.target.result}" class="w-full h-full object-cover rounded-lg"></video>
                <button type="button" data-index="${index}" onclick="removeFile(${index})" class="absolute -top-1 -right-1 bg-red-500 text-white rounded-full w-5 h-5 flex items-center justify-center text-xs hover:bg-red-600">
                  ×
                </button>
              </div>
            `;
          }
          
          preview.appendChild(previewItem);
        };
        reader.readAsDataURL(file);
      });
    }
    
    // Update preview
    updatePreview();
  };
  
  // Remove file
  window.removeFile = function(index) {
    console.log('removeFile called for index:', index);
    const input = document.getElementById('media');
    console.log('Media input:', input);
    
    if (!input) return;
    
    const dt = new DataTransfer();
    
    Array.from(input.files).forEach((file, i) => {
      if (i !== index) {
        dt.items.add(file);
      }
    });
    
    input.files = dt.files;
    console.log('Files after removal:', input.files.length);
    window.handleFileSelect(input);
  };
  
  // Clear all files
  window.clearAllFiles = function() {
    console.log('clearAllFiles called');
    const input = document.getElementById('media');
    const preview = document.getElementById('file-preview');
    
    console.log('Media input:', input);
    console.log('File preview:', preview);
    
    if (input) input.value = '';
    if (preview) preview.innerHTML = '';
    
    updatePreview();
  };
  
  // Add another post
  window.addAnotherPost = function() {
    console.log('addAnotherPost called');
    const form = document.getElementById('post-form');
    const resultDiv = document.getElementById('post-result');
    const footerContainer = document.querySelector('.p-4.border-t.border-gray-200.bg-gray-50');
    
    // Show footer
    if (footerContainer) {
      footerContainer.style.display = 'block';
    }
    
    // Reset form
    if (form) {
      const textarea = form.querySelector('#content');
      const fileInput = form.querySelector('#media');
      const datetimeInput = form.querySelector('#schedule_at_native');
      
      if (textarea) textarea.value = '';
      if (fileInput) fileInput.value = '';
      if (datetimeInput) datetimeInput.value = '';
      
      // Reset checkboxes
      const checkboxes = form.querySelectorAll('.provider-checkbox');
      checkboxes.forEach(checkbox => {
        checkbox.checked = false;
      });
    }
    
    // Clear previews
    const filePreview = document.getElementById('file-preview');
    if (filePreview) filePreview.innerHTML = '';
    
    if (resultDiv) resultDiv.innerHTML = '';
    
    // Update preview
    updatePreview();
    
    // Handle provider selection
    handleProviderSelection();
    
    // Reload providers
    loadProviders();
  };
  
  // Show preview tab
  function showPreviewTab() {
    console.log('showPreviewTab called');
    const tabContent = document.getElementById('tab-content');
    
    // Create preview content
    const previewContent = `
      <div class="mb-4">
        <h3 class="text-lg font-semibold text-gray-900 mb-3">Podgląd posta</h3>
      </div>
      <div class="space-y-4">
        <!-- Post content preview -->
        <div id="post-content-preview" class="bg-gray-50 rounded-lg p-4 min-h-[100px]">
          <p class="text-gray-500 text-sm">Treść posta pojawi się tutaj...</p>
        </div>
        
        <!-- Post media preview -->
        <div id="post-media-preview" class="space-y-2">
          <h4 class="text-sm font-medium text-gray-700">Media</h4>
          <div id="post-media-grid" class="grid grid-cols-2 gap-2">
            <!-- Media previews will be added here -->
          </div>
        </div>
      </div>
    `;
    
    tabContent.innerHTML = previewContent;
    
    // Update preview
    updatePreview();
    
    // Update tab styles
    updateTabStyles('preview');
  }
  
  // Show only preview tab
  function showOnlyPreviewTab() {
    console.log('showOnlyPreviewTab called');
    const tabContent = document.getElementById('tab-content');
    
    // Create preview content
    const previewContent = `
      <div class="mb-4">
        <h3 class="text-lg font-semibold text-gray-900 mb-3">Podgląd posta</h3>
      </div>
      <div class="space-y-4">
        <!-- Post content preview -->
        <div id="post-content-preview" class="bg-gray-50 rounded-lg p-4 min-h-[100px]">
          <p class="text-gray-500 text-sm">Treść posta pojawi się tutaj...</p>
        </div>
        
        <!-- Post media preview -->
        <div id="post-media-preview" class="space-y-2">
          <h4 class="text-sm font-medium text-gray-700">Media</h4>
          <div id="post-media-grid" class="grid grid-cols-2 gap-2">
            <!-- Media previews will be added here -->
          </div>
        </div>
      </div>
    `;
    
    tabContent.innerHTML = previewContent;
    
    // Create preview tab
    const tabsContainer = document.getElementById('provider-tabs');
    tabsContainer.innerHTML = `
        <button class="flex items-center justify-center w-10 h-10 rounded-md transition-colors bg-blue-100 text-blue-700" data-index="preview" onclick="switchToPreviewTab()" title="Preview">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
          </svg>
        </button>
    `;
    
    // Update preview
    updatePreview();
  }
  
  // Switch to preview tab
  window.switchToPreviewTab = function() {
    console.log('switchToPreviewTab called');
    showPreviewTab();
  };
  
  // Update tab styles
  function updateTabStyles(activeTabIndex) {
    const tabs = document.querySelectorAll('#provider-tabs button');
    tabs.forEach(tab => {
      const index = tab.getAttribute('data-index');
      if (index === activeTabIndex) {
        tab.classList.remove('text-gray-500', 'hover:text-gray-700');
        tab.classList.add('bg-blue-100', 'text-blue-700');
      } else {
        tab.classList.remove('bg-blue-100', 'text-blue-700');
        tab.classList.add('text-gray-500', 'hover:text-gray-700');
      }
    });
  }
  
  console.log('Post sidebar simple handler initialization complete');
});