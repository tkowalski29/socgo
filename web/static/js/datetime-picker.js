// DateTime Picker Handler
document.addEventListener('DOMContentLoaded', function() {
  console.log('DateTime picker handler loaded');
  
  // Initialize datetime picker
  function initializeDateTimePicker() {
    const datetimeInput = document.getElementById('schedule_at_footer');
    const pickerBtn = document.getElementById('datetime-picker-btn');
    const customPicker = document.getElementById('custom-datetime-picker');
    const closeBtn = document.getElementById('close-datetime-picker');
    const cancelBtn = document.getElementById('cancel-datetime');
    const applyBtn = document.getElementById('apply-datetime');
    const dateInput = document.getElementById('custom-date');
    const hourInput = document.getElementById('custom-hour');
    const minuteInput = document.getElementById('custom-minute');
    
    if (!datetimeInput || !customPicker) return;
    
    // Set default datetime
    setDefaultDateTime();
    
    // Open picker
    function openCustomPicker() {
      customPicker.style.display = 'block';
      
      // Set current values
      const currentValue = datetimeInput.value;
      if (currentValue) {
        const [datePart, timePart] = currentValue.split(' ');
        if (datePart) {
          const [year, month, day] = datePart.split('-');
          dateInput.value = `${year}-${month}-${day}`;
        }
        if (timePart) {
          const [hour, minute] = timePart.split(':');
          hourInput.value = hour;
          minuteInput.value = minute;
        }
      } else {
        // Set current date and time
        const now = new Date();
        const year = now.getFullYear();
        const month = String(now.getMonth() + 1).padStart(2, '0');
        const day = String(now.getDate()).padStart(2, '0');
        const hour = String(now.getHours()).padStart(2, '0');
        const minute = String(now.getMinutes()).padStart(2, '0');
        
        dateInput.value = `${year}-${month}-${day}`;
        hourInput.value = hour;
        minuteInput.value = minute;
      }
    }
    
    // Close picker
    function closeCustomPicker() {
      customPicker.style.display = 'none';
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
  }
  
  // Set default datetime
  function setDefaultDateTime() {
    const now = new Date();
    const year = now.getFullYear();
    const month = String(now.getMonth() + 1).padStart(2, '0');
    const day = String(now.getDate()).padStart(2, '0');
    const hour = String(now.getHours()).padStart(2, '0');
    const minute = String(now.getMinutes()).padStart(2, '0');
    
    const displayDateTime = `${year}-${month}-${day} ${hour}:${minute}`;
    const nativeDateTime = `${year}-${month}-${day}T${hour}:${minute}`;
    
    const footerScheduleInput = document.getElementById('schedule_at_footer');
    if (footerScheduleInput) footerScheduleInput.value = displayDateTime;
    
    const nativeInput = document.getElementById('schedule_at_native');
    if (nativeInput) nativeInput.value = nativeDateTime;
  }
  
  // Format datetime for display
  function formatDateTimeForDisplay(date) {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hour = String(date.getHours()).padStart(2, '0');
    const minute = String(date.getMinutes()).padStart(2, '0');
    
    return `${year}-${month}-${day} ${hour}:${minute}`;
  }
  
  // Format datetime for native input
  function formatDateTimeForNative(date) {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hour = String(date.getHours()).padStart(2, '0');
    const minute = String(date.getMinutes()).padStart(2, '0');
    
    return `${year}-${month}-${day}T${hour}:${minute}`;
  }
  
  // Initialize when DOM is ready
  initializeDateTimePicker();
  
  // Global functions
  window.setDefaultDateTime = setDefaultDateTime;
  window.formatDateTimeForDisplay = formatDateTimeForDisplay;
  window.formatDateTimeForNative = formatDateTimeForNative;
  
  // Function to open sidebar with specific date and time
  window.openPostSidebarWithDateTime = function(dateStr, hour) {
    // Check if DOM is ready
    if (document.readyState === 'loading') {
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
      openPostSidebar();
    } else {
      console.error('openPostSidebar function not found!');
      // Fallback: manually open sidebar
      sidebar.classList.add('open');
      sidebar.style.transform = 'translateY(0)';
    }
    
    // Set the datetime to the specified date and hour
    setTimeout(() => {
      const date = new Date(dateStr);
      date.setHours(hour, 0, 0, 0); // Set to the specified hour, minutes to 0
      
      const displayDateTime = formatDateTimeForDisplay(date);
      const nativeDateTime = formatDateTimeForNative(date);
      
      // Set display input
      const footerScheduleInput = document.getElementById('schedule_at_footer');
      if (footerScheduleInput) footerScheduleInput.value = displayDateTime;
      
      // Set native input for form submission
      const nativeInput = document.getElementById('schedule_at_native');
      if (nativeInput) nativeInput.value = nativeDateTime;
      
    }, 100);
  };
}); 