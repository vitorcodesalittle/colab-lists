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
          }
      }
  },
  plugins: [],
}

