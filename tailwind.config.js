const { iconsPlugin, getIconCollections } = require('@egoist/tailwindcss-icons');

/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ["./templates/**/*.html"],
    theme: {
        extend: {},
        colors: {
            brand: {
                100: '#FFF2CF',
                200: '#FFDD7E',
                300: '#FFD257',
                400: '#FFC831',
                500: '#FFBE0A',
                600: '#D79E02',
                700: '#BD8B01',
                800: '#997103'
            },
            neutral: {
                100: '#EBF5EC',
                200: '#DDE7DD',
                300: '#BDC9BE',
                400: '#AABCAB',
                500: '#9CB29D',
                600: '#849C85',
                700: '#69856A',
                800: '#EBF5EC'
            },
            ...require('tailwindcss/colors')
        },
        fontFamily: {
            'display': ["Kolker Brush"],
            'body': ["Gowun Batang"]
        }

    },
    plugins: [
        iconsPlugin({
            // Select the icon collections you want to use
            // You can also ignore this option to automatically discover all individual icon packages you have installed
            // If you install @iconify/json, you should explicitly specify the collections you want to use, like this:
            // collections: getIconCollections(["mdi", "lucide"]),
            // If you want to use all icons from @iconify/json, you can do this:
            collections: getIconCollections("all"),
            // and the more recommended way is to use `dynamicIconsPlugin`, see below.
        }),
    ],
}

