import React from "react";

class Reconfiguration extends React.Component {
  render() {
    return (
      <div className="Reconfig">
        <form name="reconfig" onSubmit={this.props.handleSubmit_CB}>
          <p>
            {" "}
            Choose the servers you want to reconfig:
            <br />
            (start the new severs first if you want to scale up)
          </p>
          <ul className="server-list">
            {this.props.serverIDs.map((server, index) => {
              return (
                <li key={index}>
                  <div className="server-list-item">
                    <input
                      className="checkbox"
                      type="checkbox"
                      id={`custom-checkbox-${index}`}
                      name={index}
                      value={server}
                      checked={this.props.checkedState[index]}
                      onChange={this.props.handleChange_CB}
                    />
                    <label htmlFor={`custom-checkbox-${index}`}>{server}</label>
                  </div>
                </li>
              );
            })}
          </ul>
          <div className="row">
            <input className="submit" type="submit" value="Submit" />
            <br />
          </div>
        </form>
        {/* {this.props.messages} */}
      </div>
    );
  }
}

export default Reconfiguration;
