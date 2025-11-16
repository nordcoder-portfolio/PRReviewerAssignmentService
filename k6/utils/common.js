export function randomChoice(arr) {
    return arr[Math.floor(Math.random() * arr.length)];
}

export function randomBetween(min, max) {
    return min + Math.random() * (max - min);
}

export function generatePrId() {
    return `pr-${__VU}-${Date.now()}-${Math.floor(Math.random() * 1e6)}`;
}

export function safeJson(res) {
    try {
        return res.json();
    } catch (e) {
        return null;
    }
}
