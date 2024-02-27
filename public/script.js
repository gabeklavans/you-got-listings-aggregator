let propertiesObject = {};
let propertiesDiv = document.getElementsByClassName("properties");

fetch("https://ygl.dabe.tech/sites")
  .then((response) => response.json())
  .then((sites) => {
    fetch("https://ygl.dabe.tech/listings")
      .then((response) => response.json())
      .then((json) => {
        propertiesObject = json;
        // console.log(propertiesObject);
        let i = 0;

        Object.entries(propertiesObject).forEach((property) => {
          let listingInfo = property[1];

          let nameTags = "";
          // find which sites the listing is from and store the names
          for (let j = 0; j < listingInfo.refs.length; j++) {
            // for each link, get the value associated with the key (the name)
            let link = listingInfo.refs[j];
            let nameLink = link.split("/rental")[0];
            let name = sites[nameLink];

            // build the link to the site around the name (<a> tag)
            nameTags += `<a href=${link} target="_blank">${name}</a>\u00A0\u00A0`;
          }

          //
          const p = $(".properties").append(
            `<p id="listing${i}" class="listing"> <span class="price">$${listingInfo.price} </span> \u00A0\u00A0\u00A0 <span class="beds-baths"> ${listingInfo.beds}Bd / ${listingInfo.baths}bth </span>  
            \u00A0\u00A0\u00A0 <span class="address">${property[0]}:</span> \u00A0\u00A0\u00A0 <span class="name-tags">${nameTags} </span> </p>
            
            <input type="text id="listing-notes" name="notes">`
          );
          i++;
        });
      });
  })

// $().add(`<p> ${property[0]} </p>`);
// console.log($());`

// bed / bath / price    LISTING    name name name

// THIS IS A SPACE IN A TEMPLATE LITERAL "  \u00A0  "
