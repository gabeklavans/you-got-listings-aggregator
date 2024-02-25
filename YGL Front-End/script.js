let propertiesObject = {};
let propertiesDiv = document.getElementsByClassName("properties");
let propertiesTextList = [];

fetch("../test.json")
  .then((response) => response.json())
  .then((json) => {
    propertiesObject = json;
    // console.log(propertiesObject);
    let i = 0;
    Object.entries(propertiesObject).forEach((property) => {
      //   let newP =
      console.log("adding new p");

      const p = $(".properties").append(
        `<p>Prop ${i + 1} .. ${property[0]} ... </p>`
      );

      propertiesTextList.push([`${property[0]}`]);
      i++;
    });
    console.log(propertiesTextList);
  });

// $().add(`<p> ${property[0]} </p>`);
// console.log($());
