import React from "react";
import styled from "styled-components";
//import "./App.css";

var client = new WebSocket("ws://127.0.0.1:12110/ws");
class App extends React.Component {
  constructor() {
    super();
    //initialize the clientID, clientSeq and accountNum
    this.state = {
      clientID: "123",
      clientSeq: 1,
      command: null,
      accountNum: 42,
      op: 0,
      amount: null,
      //ws: new WebSocket("ws://127.0.0.1:12110/ws"),
      ws: null,
      message: [],
      servers: ["127.0.0.1:12110", "127.0.0.1:12111", "127.0.0.1:12112"],
      index: 0,
      length: null,
    };
    this.handleChange = this.handleChange.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);
    //this.opNumberToString = this.opNumberToString.bind(this);
    this.state.length = this.state.servers.length;
  }
  opNumberToString(op) {
    switch (Number(op)) {
      case 0:
        return "Balance";
      case 1:
        return "Deposit";
      case 2:
        return "Withdrawl";
    }
  }
  createOpMessage(op) {
    switch (Number(op)) {
      case 0:
        return `Account number: ${this.state.accountNum}
        Operation:Balance
      `;
      case 1:
        return `Account number: ${this.state.accountNum}
        Operation:Deposit
        Amount: ${this.state.amount}
      `;
      case 2:
        return `Account number: ${this.state.accountNum}
        Operation:Withdrawl
        Amount: ${this.state.amount}
      `;
    }
  }
  //establish the websoket connection

  handleChange(event) {
    this.setState({ [event.target.name]: event.target.value }); //handle multiple event.target.value, store in an array

    console.log(this.state[event.target.name]);
  }

  handleSubmit(event) {
    let opString = this.opNumberToString(this.state.op);
    if (this.state.op == 0) {
      alert(
        "Please check the transction you are going to submit.\nThe operation: " +
          opString +
          "\nAccount number: " +
          this.state.accountNum
      );
    } else {
      alert(
        "Please check the transction you are going to submit.\nThe operation: " +
          opString +
          "\nAccount number: " +
          this.state.accountNum +
          "\nAmount: " +
          this.state.amount
      );
    }

    // alert(
    //   "Please check the transction you are going to submit.\nThe operation: " +
    //     opString +
    //     "\nAccount number: " +
    //     this.state.accountNum +
    //     "\nAmount: " +
    //     this.state.amount
    // );
    let opMessage = `Account number: ${this.state.accountNum}
      Operation:${opString}
      Amount:${this.state.amount}
    `;
    //resSting = "Account number:" + accountNum + "\nBalance:" + balance;

    opMessage = this.createOpMessage(this.state.op);
    console.log(opMessage);

    this.setState({ message: this.state.message.concat(opMessage) });

    //console.log("typeof(this.state.amount)", typeof this.state.amount);
    let op = Number(this.state.op);
    let accountNum = Number(this.state.accountNum);
    let amount = Number(this.state.amount);
    console.log("typeof(amount)", typeof amount, amount);
    let msg = {
      Command: "VALUE",
      Parameter: {
        ClientID: this.state.clientID,
        ClientSeq: this.state.clientSeq,
        Noop: false,
        AccountNum: accountNum,
        Txn: {
          Op: op,
          Amount: amount,
        },
      },
    };

    this.state.ws.send(JSON.stringify(msg));
    this.setState({ clientSeq: this.state.clientSeq + 1 });
    console.log("this.state.clientSeq", this.state.clientSeq);
    event.preventDefault();
  }

  openEventListener = (event) => {
    //send client message
    this.setState({ ws: client });
    console.log("Websocket Client Connected: ", client.url);
    let msg = {
      Command: "CLIENT",
      Parameter: {
        ClientID: this.state.clientID,
      },
    };
    this.state.ws.send(JSON.stringify(msg));
  };

  incomingMessageListener = (message) => {
    let messageData = JSON.parse(message.data);
    console.log(messageData);

    switch (messageData["Command"]) {
      case "LEADER":
        // console.log(messageData["Parameter"]);
        // console.log(
        //   "typeof(messageData[Parameter])",
        //   typeof messageData["Parameter"]
        // );
        client = new WebSocket("ws://" + messageData["Parameter"] + "/ws");
        this.setState({
          ws: new WebSocket("ws://" + messageData["Parameter"] + "/ws"),
        });
        console.log("this.state.ws.url", this.state.ws.url);
        this.wsMount();
        break;
      case "RESPONSE":
        console.log("receive response", messageData["Parameter"]["TxnRes"]);
        let accountNum = messageData["Parameter"]["TxnRes"]["AccountNum"];
        let balance = messageData["Parameter"]["TxnRes"]["Balance"];
        let resString = `Account number: ${accountNum}
        Balance:${balance}
        `;
        //resSting = "Account number:" + accountNum + "\nBalance:" + balance;
        console.log(resString);
        message = this.state.message.concat(resString);
        this.setState({ message: message });
        break;
      default:
      //console.log(messageData);
    }
    //todo:message dispatching
  };

  closeSocket = (event) => {
    console.log("You are disconnected");
    //todo:reconnect
    /*
    this.newWS();
    setTimeout(
      function () {
        console.log("this.state.ws.url", this.state.ws.url);
        this.wsMount();
      }.bind(this),
      50
    );*/
    //give up after a cetain times of attempts of re-connention
    if (this.state.index < 8) {
      setTimeout(() => {
        console.log("Retrying connection times:", this.state.index);
        client = new WebSocket(
          "ws://" +
            this.state.servers[this.state.index % this.state.length] +
            "/ws"
        );
        this.setState({ index: this.state.index + 1 });
        client.addEventListener("open", this.openEventListener);
        client.addEventListener("message", this.incomingMessageListener);
        client.addEventListener("close", this.closeSocket);
      }, 5000);
    }
  };

  wsMount() {
    console.log("wsMount()");
    /*
    this.state.ws.addEventListener("open", this.openEventListener);
    this.state.ws.addEventListener("message", this.incomingMessageListener);
    this.state.ws.addEventListener("close", this.closeSocket);
    */
    client.addEventListener("open", this.openEventListener);
    client.addEventListener("message", this.incomingMessageListener);
    client.addEventListener("close", this.closeSocket);
  }

  componentDidMount() {
    console.log("componentDidMount()");
    //this.newWS();
    //console.log("client.readyState", this.state.ws.readyState);
    //console.log("client:", this.state.ws);
    this.wsMount();
  }

  render() {
    const messages = this.state.message.map((msg, index) => (
      <p key={index}>{msg}</p>
    ));
    return (
      <AppStyled>
        <div className="App">
          <h2>Welcome to Bank One</h2>
          <form onSubmit={this.handleSubmit}>
            <div className="row">
              <label>
                Choose your transction operation:
                {/*controlled element: the value attribute is set on our form element, the displayed value 
            will always be this.state.value, making the React state the source of truth.  */}
                <select
                  name="op" //name is same as the state name
                  value={this.state.op}
                  onChange={this.handleChange}
                >
                  <option value="0">Balance</option>
                  <option value="1">Deposit</option>
                  <option value="2">Withdrawal</option>
                </select>
              </label>
              <br />
            </div>
            <div className="row">
              <label>
                Select the account number:
                {/*controlled element: the value attribute is set on our form element, the displayed value 
            will always be this.state.value, making the React state the source of truth.  */}
                <select
                  name="accountNum" //name is same as the state name
                  value={this.state.accountNum}
                  onChange={this.handleChange}
                >
                  <option value="42">42</option>
                  <option value="52">52</option>
                  <option value="62">62</option>
                </select>
              </label>
              <br />
            </div>
            <div className="row">
              <label>
                Please input your amount:
                <input
                  className="amount"
                  name="amount"
                  type="number"
                  value={this.state.amount}
                  onChange={this.handleChange}
                />
              </label>
              <div className="row">
                <input className="submit" type="submit" value="Submit" />
                <br />
              </div>
            </div>
          </form>
          {messages}
        </div>
      </AppStyled>
    );
  }
}

const AppStyled = styled.div`
  position: absolute;
  left: 35%;
  .row {
    margin-top: 5px;
    select {
      margin-left: 5px;
    }
    .amount {
      margin-left: 5px;
    }
  }
`;
export default App;
