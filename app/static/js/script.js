// User selections
const dislikedIngredients = new Set();
const preferredIngredients = new Set();
const apiBaseUrl = '/api'; // Use relative path for API calls via Nginx proxy

// Function to populate ingredient lists from API
async function populateIngredientLists() {
    try {
        const response = await fetch(`${apiBaseUrl}/ingredients`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const ingredients = await response.json();

        const dislikedList = document.getElementById('all-ingredients-disliked');
        const preferredList = document.getElementById('all-ingredients-preferred');
        dislikedList.innerHTML = ''; // Clear any placeholders
        preferredList.innerHTML = ''; // Clear any placeholders

        ingredients.forEach(ingredient => {
            // Create chip for disliked list
            const dislikedChip = document.createElement('div');
            dislikedChip.className = 'ingredient-chip';
            dislikedChip.setAttribute('data-ingredient', ingredient);
            dislikedChip.textContent = ingredient;
            dislikedChip.addEventListener('click', function() { toggleDisliked(ingredient, this); });
            dislikedList.appendChild(dislikedChip);

            // Create chip for preferred list
            const preferredChip = document.createElement('div');
            preferredChip.className = 'ingredient-chip';
            preferredChip.setAttribute('data-ingredient', ingredient);
            preferredChip.textContent = ingredient;
            preferredChip.addEventListener('click', function() { togglePreferred(ingredient, this); });
            preferredList.appendChild(preferredChip);
        });

    } catch (error) {
        console.error('Error fetching or populating ingredients:', error);
        // Optionally display an error message to the user on the page
        document.getElementById('all-ingredients-disliked').textContent = 'Error loading ingredients.';
        document.getElementById('all-ingredients-preferred').textContent = 'Error loading ingredients.';
    }
}


// Setup ingredient selection - **MOVED EVENT LISTENER SETUP INSIDE populateIngredientLists**
// document.querySelectorAll('#all-ingredients-disliked .ingredient-chip').forEach(chip => { ... });
// document.querySelectorAll('#all-ingredients-preferred .ingredient-chip').forEach(chip => { ... });

// Toggle disliked ingredient
function toggleDisliked(ingredient, chipElement) {
    if (dislikedIngredients.has(ingredient)) {
        dislikedIngredients.delete(ingredient);
        chipElement.classList.remove('disliked');
        removeFromSelectedDisplay(ingredient, 'disliked');
    } else {
        if (preferredIngredients.has(ingredient)) {
            // Remove from preferred if it's there
            preferredIngredients.delete(ingredient);
            document.querySelectorAll(`#all-ingredients-preferred .ingredient-chip[data-ingredient="${ingredient}"]`).forEach(chip => {
                chip.classList.remove('preferred');
            });
            removeFromSelectedDisplay(ingredient, 'preferred');
        }
        
        dislikedIngredients.add(ingredient);
        chipElement.classList.add('disliked');
        addToSelectedDisplay(ingredient, 'disliked');
    }
}

// Toggle preferred ingredient
function togglePreferred(ingredient, chipElement) {
    if (preferredIngredients.has(ingredient)) {
        preferredIngredients.delete(ingredient);
        chipElement.classList.remove('preferred');
        removeFromSelectedDisplay(ingredient, 'preferred');
    } else {
        if (dislikedIngredients.has(ingredient)) {
            // Remove from disliked if it's there
            dislikedIngredients.delete(ingredient);
            document.querySelectorAll(`#all-ingredients-disliked .ingredient-chip[data-ingredient="${ingredient}"]`).forEach(chip => {
                chip.classList.remove('disliked');
            });
            removeFromSelectedDisplay(ingredient, 'disliked');
        }
        
        preferredIngredients.add(ingredient);
        chipElement.classList.add('preferred');
        addToSelectedDisplay(ingredient, 'preferred');
    }
}

// Add ingredient to selected display
function addToSelectedDisplay(ingredient, type) {
    const container = document.getElementById(`${type}-ingredients`);
    const chip = document.createElement('span');
    chip.className = `selected-chip ${type}-chip`;
    chip.textContent = ingredient;
    container.appendChild(chip);
}

// Remove ingredient from selected display
function removeFromSelectedDisplay(ingredient, type) {
    const container = document.getElementById(`${type}-ingredients`);
    container.querySelectorAll('.selected-chip').forEach(chip => {
        if (chip.textContent === ingredient) {
            container.removeChild(chip);
        }
    });
}

// Get recommendations from the API
document.getElementById('recommend-btn').addEventListener('click', async function() {
    const loadingElement = document.getElementById('loading');
    const resultsElement = document.getElementById('results');
    
    // Show loading, hide results
    loadingElement.style.display = 'block';
    resultsElement.innerHTML = '';
    
    try {
        // Use the relative API base URL
        const response = await fetch(`${apiBaseUrl}/recommend`, { // Updated URL
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                disliked: Array.from(dislikedIngredients),
                preferred: Array.from(preferredIngredients)
            })
        });
        
        const recommendations = await response.json();
        
        // Hide loading
        loadingElement.style.display = 'none';
        
        if (recommendations.length === 0) {
            resultsElement.innerHTML = '<div class="no-results">No matching pizzas found. Try adjusting your preferences.</div>';
            return;
        }
        
        // Display recommendations
        recommendations.forEach(pizza => {
            const pizzaCard = document.createElement('div');
            pizzaCard.className = 'pizza-card';
            
            const pizzaImage = document.createElement('div');
            pizzaImage.className = 'pizza-image';
            pizzaImage.textContent = '🍕';
            
            const pizzaName = document.createElement('div');
            pizzaName.className = 'pizza-name';
            pizzaName.textContent = pizza.name;
            
            const pizzaIngredients = document.createElement('div');
            pizzaIngredients.className = 'pizza-ingredients';
            pizzaIngredients.textContent = pizza.ingredients.join(', ');
            
            pizzaCard.appendChild(pizzaImage);
            pizzaCard.appendChild(pizzaName);
            pizzaCard.appendChild(pizzaIngredients);
            resultsElement.appendChild(pizzaCard);
        });
        
        // Scroll to results
        resultsElement.scrollIntoView({ behavior: 'smooth' });
        
    } catch (error) {
        console.error('Error fetching recommendations:', error);
        loadingElement.style.display = 'none';
        resultsElement.innerHTML = '<div class="no-results">An error occurred fetching recommendations. Please try again or check the console.</div>'; 
    }
}); 

// Fetch ingredients when the DOM is fully loaded
document.addEventListener('DOMContentLoaded', populateIngredientLists); 