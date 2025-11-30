# Manual Test: Sticky Header Feature

## Test Date: November 30, 2025
## Feature: Sticky header that stays fixed at top and shrinks on scroll

---

## Prerequisites
- Application running: `docker compose up -d`
- Browser: Chrome, Firefox, or Safari
- URL: http://localhost:8080

---

## Test Cases

### 1. Initial Load ✓
**Steps:**
1. Open http://localhost:8080
2. Observe header appearance

**Expected:**
- Header visible at top
- Full-size search bar (2.5rem height)
- Normal padding (0.5rem)
- Light shadow
- "Demo Store" brand text normal size

**Status:** [ ] PASS [ ] FAIL

---

### 2. Scroll Down (>50px) ✓
**Steps:**
1. Scroll down the page past 50 pixels

**Expected:**
- Header remains fixed at top (doesn't scroll away)
- Header smoothly shrinks (0.3s transition)
- Search bar height reduces to 2rem
- Padding reduces to 0.25rem
- Brand text reduces to 1.2rem
- Shadow increases (4px with backdrop blur)
- Background becomes semi-transparent (95% opacity)

**Status:** [ ] PASS [ ] FAIL

---

### 3. Scroll Back to Top ✓
**Steps:**
1. Scroll back to the very top of page

**Expected:**
- Header smoothly expands back to original size
- All elements return to original dimensions
- Transitions are smooth (no jerky movement)

**Status:** [ ] PASS [ ] FAIL

---

### 4. Interactive Elements While Scrolled ✓
**Steps:**
1. Scroll down page
2. Hover over "Cart" button
3. Click search bar
4. Type in search bar
5. Click account/menu dropdown

**Expected:**
- Cart dropdown opens on hover
- Search bar is clickable and functional
- Typing works in search bar
- Menus open and close correctly
- All elements remain accessible
- No z-index conflicts

**Status:** [ ] PASS [ ] FAIL

---

### 5. Content Not Hidden ✓
**Steps:**
1. Check that page content starts below header
2. Scroll through entire page

**Expected:**
- First product/content is visible (not hidden under header)
- Body has 80px padding-top
- No content overlap with fixed header
- All page content scrolls normally

**Status:** [ ] PASS [ ] FAIL

---

### 6. Mobile Responsive (Width < 991px) ✓
**Steps:**
1. Resize browser window to < 991px width
2. Scroll down and up

**Expected:**
- Body padding-top reduces to 70px (mobile)
- Search bar hidden (display: none)
- Hamburger menu (☰) visible
- Sticky behavior still works
- Header shrinks appropriately on scroll

**Status:** [ ] PASS [ ] FAIL

---

### 7. Multiple Scroll Events ✓
**Steps:**
1. Rapidly scroll up and down multiple times
2. Observe header behavior

**Expected:**
- No flickering or glitches
- Smooth transitions every time
- Class toggles work reliably
- No JavaScript errors in console

**Status:** [ ] PASS [ ] FAIL

---

### 8. Cross-Browser Compatibility ✓
**Steps:**
1. Test in Chrome
2. Test in Firefox
3. Test in Safari (if available)

**Expected:**
- Sticky positioning works in all browsers
- Backdrop-filter blur works (or gracefully degrades)
- Transitions smooth in all browsers
- No visual differences in layout

**Status:**
- Chrome: [ ] PASS [ ] FAIL
- Firefox: [ ] PASS [ ] FAIL
- Safari: [ ] PASS [ ] FAIL

---

## Visual Inspection Checklist

- [ ] No horizontal scrollbar appears
- [ ] Header width matches container
- [ ] No white gaps or spacing issues
- [ ] Shadow effect visible and appropriate
- [ ] Text remains readable when scrolled
- [ ] Cart badge visible and correct
- [ ] All links and buttons clickable

---

## JavaScript Console Check

Open browser DevTools (F12) → Console tab

**Check for:**
- [ ] No JavaScript errors
- [ ] No CSS-related warnings
- [ ] Scroll event handler works (test by logging)

---

## Performance Check

**Observe:**
- [ ] Smooth scrolling (no lag)
- [ ] No choppy animations
- [ ] Header transition feels responsive
- [ ] Page doesn't feel "heavy"

---

## Known Issues / Notes

- Backdrop-filter may not work in older browsers (graceful degradation)
- iOS Safari may have different fixed positioning behavior

---

## Test Result

**Overall Status:** [ ] ALL PASS [ ] SOME FAILED

**Tester:** _________________

**Notes:**

---

## If Tests Fail

1. Check browser console for errors
2. Verify CSS classes being applied (use DevTools → Elements)
3. Check if JavaScript scroll listener is firing (add console.log)
4. Verify z-index values (should be 1000 for header)
5. Check for CSS conflicts with Pico.css


