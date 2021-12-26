import React from "react";
import styled from "styled-components";

class Service extends React.Component {
  //   constructor() {
  //     super();
  //     //initialize the clientID, clientSeq and accountNum
  //   }

  render() {
    return (
      <div className="Service">
        <form onSubmit={this.props.handleSubmit}>
          <div className="row">
            <label>
              Choose your transction operation:
              {/*controlled element: the value attribute is set on our form element, the displayed value 
            will always be this.state.value, making the React state the source of truth.  */}
              <select
                name="op" //name is same as the state name
                value={this.props.op}
                onChange={this.props.handleChange}
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
                value={this.props.accountNum}
                onChange={this.props.handleChange}
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
                value={this.props.amount}
                onChange={this.props.handleChange}
              />
            </label>
            <div className="row">
              <input className="submit" type="submit" value="Submit" />
              <br />
            </div>
          </div>
        </form>
        {this.props.messages}
      </div>
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
export default Service;
