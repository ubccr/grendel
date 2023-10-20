/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ["./frontend/views/**/*.gohtml"],
    theme: {
        extend: {
            keyframes: {
                loading: {
                    "0%": {
                        left: "0%",
                        right: "100%",
                        width: "0%",
                    },
                    "10%": {
                        left: "0%",
                        right: "75%",
                        width: "25%",
                    },
                    "90%": {
                        right: "0%",
                        left: "75%",
                        width: "25%",
                    },
                    "100%": {
                        left: "100%",
                        right: "0%",
                        width: "0%",
                    },
                },
                wiggle: {
                    "0%": {
                        transform: "rotate(-0.5deg)",
                    },
                    "50%": {
                        transform: "rotate(0.7deg)",
                    },
                },
            },
            animation: {
                loading: "loading 2s linear infinite",
                wiggle: "wiggle 0.25s ease-in-out infinite",
            },
            backgroundImage: {
                "large-triangles-ub": "url('/backgrounds/large-triangles-ub.svg')",
                "large-triangles-dark": "url('/backgrounds/large-triangles-dark.svg')",
                "large-triangles": "url('/backgrounds/large-triangles.svg')",
                "protruding-squares-ub": "url('/backgrounds/protruding-squares-ub.svg')",
                "protruding-squares": "url('/backgrounds/protruding-squares.svg')",
            },
        },
    },
    plugins: [],
};
