import { initCanvas } from './canvas.js';

document.addEventListener('DOMContentLoaded', () => {
  // --- GENERATIVE BACKGROUND ---
  initCanvas('zero-egress-canvas');

  // --- STANDALONE ROI CALCULATOR LOGIC ---
  const dataInput = document.getElementById('data-volume');
  const discountInput = document.getElementById('cloud-discount');
  const cloudCostDisplay = document.getElementById('cloud-total-cost');
  const savingsDisplay = document.getElementById('annual-savings');

  const OCULTAR_MONTHLY = 10000;
  
  const formatter = new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    maximumFractionDigits: 0
  });

  function updateROI() {
    if (!dataInput || !discountInput) return;

    const requests = parseFloat(dataInput.value) || 0;
    const discount = parseFloat(discountInput.value) || 0;
    
    // Base cost: $5.00 per 10k requests (hypothetical enterprise pricing)
    const baseRate = 0.0005; 
    const monthlyCost = requests * baseRate * (1 - (discount/100));
    const annualSavings = (monthlyCost - OCULTAR_MONTHLY) * 12; 
    
    if (cloudCostDisplay) cloudCostDisplay.textContent = formatter.format(monthlyCost);
    if (savingsDisplay) savingsDisplay.textContent = formatter.format(annualSavings > 0 ? annualSavings : 0);
  }

  if (dataInput && discountInput) {
    dataInput.addEventListener('input', updateROI);
    discountInput.addEventListener('input', updateROI);
    updateROI();
  }

  // --- SMOOTH SCROLLING ---
  document.querySelectorAll('a[href^="#"]').forEach(anchor => {
    anchor.addEventListener('click', function (e) {
      const href = this.getAttribute('href');
      if (href === '#' || !href.startsWith('#')) return;
      e.preventDefault();
      const target = document.querySelector(href);
      if (target) {
        target.scrollIntoView({
          behavior: 'smooth'
        });
      }
    });
  });

  console.log("OCULTAR_SYSTEM_v1.0.4: BOOT_COMPLETE");
});
