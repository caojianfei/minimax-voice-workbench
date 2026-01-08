<script setup>
import { computed } from 'vue'

const currentYear = computed(() => new Date().getFullYear())
</script>

<template>
  <footer class="app-footer">
    <div class="footer-content">
      <div class="copyright">
        &copy; {{ currentYear }} Minimax Voice Workbench. All rights reserved.
      </div>
      <div class="links">
        <a href="#" class="footer-link">Privacy Policy</a>
        <span class="divider">•</span>
        <a href="#" class="footer-link">Terms of Service</a>
      </div>
    </div>
  </footer>
</template>

<style scoped>
.app-footer {
  margin-top: auto;
  padding: var(--space-6) 0;
  border-top: 1px solid var(--border-color);
  color: var(--text-tertiary);
  font-size: 0.875rem;
}

.footer-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
}

@media (min-width: 768px) {
  .footer-content {
    flex-direction: row;
    justify-content: space-between;
  }
}

.links {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.footer-link {
  color: var(--text-secondary);
  text-decoration: none;
  transition: color var(--transition-fast);
}

.footer-link:hover {
  color: var(--primary);
}

.divider {
  color: var(--border-color);
}
</style>
