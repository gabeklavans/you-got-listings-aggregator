const BrokerKeys = [
  "BrokerNames",
  "BrokerUrls",
]

const FilterKeys = [
  "BedsMin",
  "BedsMax",
  "BathsMin",
  "BathsMax",
  "RentMin",
  "RentMax",
  "DateMin",
  "DateMax",
  "ExcludedAreas",
]

const NotificationKeys = [
  "NotificationUrls"
]

function minMaxBoundsCheck(minElement, maxElement) {
  const elements = [minElement, maxElement];
  for (const element of elements) {
    element.addEventListener("input", _ => {
      if (parseInt(minElement.value.replaceAll("-", "")) > parseInt(maxElement.value.replaceAll("-", ""))) {
        element.setCustomValidity("Maximum cannot be greater than minimum.")
      } else {
        element.setCustomValidity("")
      }
    });
  }
}

function matchingCommaCheck(nameElement, urlElement) {
  const elements = [nameElement, urlElement]
  for (const element of elements) {
    element.addEventListener("input", _ => {
      if (!nameElement.value && !urlElement.value) {
        return;
      } else if ((!nameElement.value || !urlElement.value) ||
        (nameElement.value.split(",").length !== urlElement.value.split(",").length)) {
        elements.forEach(e => e.setCustomValidity("Broker settings must both have the same number of commas."));
      } else {
        elements.forEach(e => e.setCustomValidity(""));
      }
    })
  }
}

minMaxBoundsCheck(document.getElementsByName("BedsMin")[0], document.getElementsByName("BedsMax")[0]);
minMaxBoundsCheck(document.getElementsByName("BathsMin")[0], document.getElementsByName("BathsMax")[0]);
minMaxBoundsCheck(document.getElementsByName("RentMin")[0], document.getElementsByName("RentMax")[0]);
minMaxBoundsCheck(document.getElementsByName("DateMin")[0], document.getElementsByName("DateMax")[0]);
matchingCommaCheck(document.getElementsByName("BrokerNames")[0], document.getElementsByName("BrokerUrls")[0]);

settingsForm = document.forms["settings"];
settingsForm.addEventListener(
  "submit",
  async (event) => {
    event.preventDefault();
    const formData = new FormData(
      settingsForm,
      document.querySelector("button[value=Save]"),
    );

    let brokerNames = []
    let brokerUrls = []
    const brokers = []
    const filters = []
    let notifications = []
    for (const [key, value] of formData) {
      if (!value) {
        continue;
      }
      console.log(key)

      if (BrokerKeys.includes(key)) {
        // store the broker info in separate lists due to comma-parsing
        if (key == "BrokerNames") {
          brokerNames = value.split(",")
        } else if (key == "BrokerUrls") {
          brokerUrls = value.split(",")
        }
      } else if (FilterKeys.includes(key)) {
        filters.push({
          name: key,
          value: value,
        });
      } else if (NotificationKeys.includes(key)) {
        // split the notification URIs by comma into the proper JSON object structure
        notifications = value.split(",").map((notifUri) => {
          return {
            url: notifUri
          };
        });
      }
    }

    // build the brokers list using the keys and urls parsed from before
    for (let idx = 0; idx < brokerNames.length; idx++) {
      brokers.push({
        name: brokerNames[idx],
        url: brokerUrls[idx]
      })
    }
    console.debug("sending brokers");
    console.debug(brokers)
    await fetch(`${window.location.origin}/v1/brokers`, {
      method: "PATCH",
      body: JSON.stringify(brokers)
    });

    console.debug("sending filters");
    console.debug(filters)
    await fetch(`${window.location.origin}/v1/filters`, {
      method: "PATCH",
      body: JSON.stringify(filters)
    });

    console.debug("sending notifications");
    console.debug(notifications)
    await fetch(`${window.location.origin}/v1/notifications`, {
      method: "PATCH",
      body: JSON.stringify(notifications)
    });

    alert(`I guess I'll filter on those.... not because I want to or anything...... >.<'`);

    document.location.assign(document.location.origin);
  },
  false,
);
