document.addEventListener('DOMContentLoaded', () => {
  // DOM Elements
  const searchInput = document.getElementById('student-search-input');
  const dropdown = document.getElementById('autocomplete-dropdown');
  const selectedPill = document.getElementById('selected-student-pill');
  const clearPillBtn = document.getElementById('clear-student-btn');
  const pillNameSpan = selectedPill?.querySelector('.student-pill-name');

  const stepDetails = document.getElementById('step-details');
  const stepVerify = document.getElementById('step-verify');

  const attendanceCards = document.querySelectorAll('.attendance-card');
  const foodPreferenceSection = document.getElementById('food-preference-section');
  const foodCards = document.querySelectorAll('.food-card');

  const emailMasked = document.getElementById('student-email-masked');
  const submitBtn = document.getElementById('rsvp-submit-btn');
  const rsvpForm = document.getElementById('rsvp-form');

  const statusOverlay = document.getElementById('rsvp-status-overlay');
  const successOverlay = document.getElementById('rsvp-success-overlay');

  // Application State
  let state = {
    selectedStudent: null,
    attendance: null,       // 'confirmed' | 'declined'
    foodPreference: 'veg',  // 'veg' | 'non-veg'
  };

  // Debouncing helper
  let debounceTimeout = null;

  // Mask email utility (e.g., reetamshyamal2005@gmail.com -> ree***@gmail.com)
  function maskEmail(email) {
    if (!email) return '';
    const parts = email.split('@');
    if (parts.length !== 2) return email;
    const name = parts[0];
    const domain = parts[1];
    
    if (name.length <= 3) {
      return `${name.slice(0, 1)}***@${domain}`;
    }
    return `${name.slice(0, 3)}***@${domain}`;
  }

  // Auto-suggest student name search
  if (searchInput) {
    searchInput.addEventListener('input', () => {
      clearTimeout(debounceTimeout);
      const query = searchInput.value.trim();

      if (query.length < 2) {
        hideDropdown();
        return;
      }

      debounceTimeout = setTimeout(() => {
        fetch(`/api/students?query=${encodeURIComponent(query)}`)
          .then(res => res.json())
          .then(data => {
            renderSuggestions(data);
          })
          .catch(err => {
            console.error('Failed to fetch students:', err);
          });
      }, 300);
    });

    // Close dropdown on click outside
    document.addEventListener('click', (e) => {
      if (!searchInput.contains(e.target) && !dropdown.contains(e.target)) {
        hideDropdown();
      }
    });
  }

  function renderSuggestions(students) {
    if (!dropdown) return;
    dropdown.innerHTML = '';

    if (!students || students.length === 0) {
      dropdown.innerHTML = '<div class="autocomplete-item scribble" style="text-align: center; color: var(--rose-burgundy);">"No seniors match your query..."</div>';
      showDropdown();
      return;
    }

    students.forEach(student => {
      const item = document.createElement('div');
      item.className = 'autocomplete-item';
      item.textContent = student.name;
      item.addEventListener('click', () => selectStudent(student));
      dropdown.appendChild(item);
    });

    showDropdown();
  }

  function showDropdown() {
    dropdown?.classList.remove('hidden');
  }

  function hideDropdown() {
    dropdown?.classList.add('hidden');
  }

  // Select student card
  function selectStudent(student) {
    state.selectedStudent = student;
    hideDropdown();

    if (searchInput) searchInput.value = '';
    if (searchInput) searchInput.classList.add('hidden');

    if (pillNameSpan) pillNameSpan.textContent = `Selected: ${student.name}`;
    selectedPill?.classList.remove('hidden');

    // Display masked email in Step 3
    if (emailMasked) emailMasked.textContent = maskEmail(student.email);

    // Unlock Step 2
    stepDetails?.classList.remove('locked');
  }

  // Clear selected student
  if (clearPillBtn) {
    clearPillBtn.addEventListener('click', () => {
      state.selectedStudent = null;
      state.attendance = null;
      state.foodPreference = 'veg';

      // Reset UI selections
      attendanceCards.forEach(c => c.classList.remove('active'));
      foodCards.forEach(c => c.classList.remove('active'));
      document.querySelector('.food-card[data-val="veg"]')?.classList.add('active');
      foodPreferenceSection?.classList.add('hidden');

      selectedPill?.classList.add('hidden');
      if (searchInput) {
        searchInput.classList.remove('hidden');
        searchInput.focus();
      }

      // Lock subsequent steps
      stepDetails?.classList.add('locked');
      stepVerify?.classList.add('locked');
      if (submitBtn) submitBtn.disabled = true;
    });
  }

  // Attendance card selection
  attendanceCards.forEach(card => {
    card.addEventListener('click', () => {
      if (stepDetails?.classList.contains('locked')) return;

      attendanceCards.forEach(c => c.classList.remove('active'));
      card.classList.add('active');

      const attendanceVal = card.getAttribute('data-val');
      state.attendance = attendanceVal;

      if (attendanceVal === 'confirmed') {
        foodPreferenceSection?.classList.remove('hidden');
      } else {
        foodPreferenceSection?.classList.add('hidden');
      }

      // Unlock Step 3
      stepVerify?.classList.remove('locked');
      if (submitBtn) submitBtn.disabled = false;
    });
  });

  // Food card selection
  foodCards.forEach(card => {
    card.addEventListener('click', () => {
      if (foodPreferenceSection?.classList.contains('hidden')) return;

      foodCards.forEach(c => c.classList.remove('active'));
      card.classList.add('active');

      state.foodPreference = card.getAttribute('data-val');
    });
  });

  // Submit RSVP Form
  if (rsvpForm) {
    rsvpForm.addEventListener('submit', (e) => {
      e.preventDefault();

      if (!state.selectedStudent || !state.attendance) return;

      // Show loader
      statusOverlay?.classList.remove('hidden');

      const payload = {
        studentId: state.selectedStudent.id,
        rsvpStatus: state.attendance,
        foodPreference: state.attendance === 'confirmed' ? state.foodPreference : ''
      };

      fetch('/api/rsvp', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(payload)
      })
      .then(async res => {
        const data = await res.json();
        if (!res.ok) {
          throw new Error(data.error || 'Server returned an error');
        }
        return data;
      })
      .then(data => {
        // Show success overlay
        statusOverlay?.classList.add('hidden');
        successOverlay?.classList.remove('hidden');
      })
      .catch(err => {
        statusOverlay?.classList.add('hidden');
        alert(`RSVP Error: ${err.message}`);
      });
    });
  }
});
