let propertiesObject = {};
let propertiesDiv = document.getElementsByClassName("properties");

const dateStringOptions = {
  weekday: "long",
  year: "numeric",
  month: "long",
  day: "numeric",
}

// see index.tmpl for where these consts come from
fetch(`${BASE_URL}/v1/sites`)
  .then((response) => response.json())
  .then((sites) => {
    fetch(`${BASE_URL}/v1/listings`)
      .then((response) => response.json())
      .then((propertiesObject) => {
        // console.log(propertiesObject);
        let i = 0;

        Object.entries(propertiesObject)
            .sort((a, b) => b[1].timestamp - a[1].timestamp) // desc order by time
            .forEach((property) => {
          let listingInfo = property[1];

          let nameTags = "";
          refs = listingInfo.refs.split(",")
          // find which sites the listing is from and store the names
          for (let j = 0; j < refs.length; j++) {
            // for each link, get the value associated with the key (the name)
            let link = refs[j];
            let nameLink = link.split("/rental")[0];
            let name = sites[nameLink];

            // build the link to the site around the name (<a> tag)
            nameTags += `<a href=${link} target="_blank">${name}</a>\u00A0\u00A0`;
          }

          //
          dateObj = new Date(listingInfo.timestamp / 1000000)
          const p = $(".properties").append(
            `<p id="listing${i}" class="listing"> <span class="timestamp">(${dateObj.toLocaleDateString("en-US", dateStringOptions)} ${dateObj.toLocaleTimeString("en-US")})\u00A0 - \u00A0</span> <span class="price">$${listingInfo.price} </span> \u00A0\u00A0 <span class="beds-baths"> ${listingInfo.beds}Bd / ${listingInfo.baths}bth </span>  
            \u00A0\u00A0\u00A0 <span class="address">${property[0]}:</span> \u00A0\u00A0\u00A0 <span class="name-tags">${nameTags} </span> </p>
            
            <input type="text id="listing-notes" name="notes">`
          );
          i++;
        });
      });
  });

// THIS IS A SPACE IN A TEMPLATE LITERAL "  \u00A0  "
