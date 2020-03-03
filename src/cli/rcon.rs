use serde_derive::Deserialize;
use clap::{Arg, App};
use std::path::Path;
use std::fs::File;
use std::io::Read;
use std::error::Error;
use std::process::Command;
use std::fmt;
use toml::de;
use std::io;

#[derive(Debug)]
struct RconError {
    msg: String,
}

impl RconError {
    fn new(msg: &str) -> RconError {
        RconError{msg: msg.to_string()}
    }
}

impl fmt::Display for RconError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.msg)
    }
}

impl Error for RconError {
    fn description(&self) -> &str {
        &self.msg
    }
}

impl From<io::Error> for RconError {
    fn from(err: io::Error) -> Self {
        RconError::new(&std::format!("I/O Error: {}", err.description()))
    }
}

impl From<de::Error> for RconError {
    fn from(err: de::Error) -> Self {
        RconError::new(&std::format!("Deserialize Error: {}", err.description()))
    }
}

#[derive(Deserialize)]
/// Specification for a run context
struct Config {
    /// Full path to the executable of the run target
    run: String,
    /// Optionally specify the current work directory the process should start in
    wd: Option<String>,
}

// The run context. It will be built up via Builder pattern
struct RunContext {
    cfg: Config
}

impl RunContext {
    fn run(&self) -> i32 {
        println!("Running {}.", self.cfg.run);
        let mut cmd = Command::new(&self.cfg.run);
        if let Some(d) = &self.cfg.wd {
            cmd.current_dir(d);
        };
        let output = cmd.output();
        match output {
            Ok(_r) => 0,
            Err(_e) => 1,
        }
    }
}

fn main() {
    let args = App::new("rcon - run context")
                        .version("0.1")
                        .author("Scott Hussey @sthussey")
                        .about("Build runtime contexts for application testing.")
                        .arg(Arg::with_name("FILE")
                            .short("f")
                            .long("file")
                            .value_name("FILE")
                            .required(true)
                            .help("TOML file specifying the run context.")
                            .takes_value(true)).get_matches();

    let config = load_config(args.value_of("FILE").unwrap());

    let runctx = match config {
        Ok(config) => RunContext { cfg: config },
        Err(err) => {
            println!("Error parsing config: {}", err);
            std::process::exit(1);
        },
    };

    std::process::exit(runctx.run());
}

// Find, read, and parse a config file
fn load_config(file_path: &str) -> Result<Config, RconError> {
    let path = Path::new(file_path);
    let mut file = File::open(&path)?;
    let mut contents = Vec::new();
    file.read_to_end(&mut contents)?;
    let config = toml::from_slice(&contents);
    match config {
        Ok(c) => Ok(c),
        Err(e) => Err(RconError::from(e)),
    }
}

#[cfg(test)]
mod tests {
    #[test]
    fn config_file_not_found() -> Result<(), String> {
        let cfg = super::load_config("foo");
        match cfg {
            Err(e) => println!("load_config erred on file not found."),
            Ok(v) => return Err(String::from("load_config did not catch file not found.")),
        };
        Ok(())
    }
}
