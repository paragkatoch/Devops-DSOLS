import http from 'k6/http';

// export let options = {
//     vus: 40,
//     duration: '60s',
// };

export let options = {
    stages: [
        // { duration: '1m', target: 20 },
        { duration: '1m', target: 40 },
        { duration: '1m', target: 45 },
        { duration: '1m', target: 50 },
    ],
};

// export let options = {
//     scenarios: {
//         test: {
//             executor: 'constant-arrival-rate',
//             rate: 500, // 500 requests/sec
//             timeUnit: '1s',
//             duration: '1m',
//             preAllocatedVUs: 100,
//             maxVUs: 500,
//         },
//     },
// };

// export let options = {
//     stages: [
//         { duration: '1m', target: 25 },
//         { duration: '1m', target: 50 },
//         { duration: '1m', target: 75 },
//         { duration: '1m', target: 100 },
//         // { duration: '1m', target: 200 },

//         { duration: '1m', target: 300 },
//         { duration: '1m', target: 400 },
//         { duration: '1m', target: 500 },
//     ],
// };

// export let options = {
//     scenarios: {
//         test: {
//             executor: 'constant-arrival-rate',
//             rate: 800, // start lower, then increase
//             timeUnit: '1s',
//             duration: '2m',
//             preAllocatedVUs: 100,
//             maxVUs: 500,
//         },
//     },
// };


function getRandomInt() {
    return Math.floor(Math.random() * 10) - 4;
}

function getRandomSwitch(rg) {
    return Math.floor(Math.random() * rg);
}

let address = "127.0.0.1"

export default function () {
    if (getRandomSwitch(2) < 2) {
        const payload = JSON.stringify({
            id: getRandomSwitch(2) ? "1" : "2",
            quantity: getRandomInt(),
        });

        const params = {
            headers: {
                'Content-Type': 'application/json',
            },
        };

        http.post(
            "http://127.0.0.1/api/product/quantity",
            payload,
            params
        );
    } else {
        http.get('http://127.0.0.1/api/product');
    }

}