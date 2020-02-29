use serde_derive::Deserialize;
use clap::{Arg, App};
use std::path::Path;
use std::fs::File;
use std::io::Read;
use toml::de;
use std::io;

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

fn main() {
    let args = App::new("rcon - run context")
                        .version("0.1")
                        .author("Scott Hussey @sthussey")
                        .about("Build runtime contexts for application testing.")
                        .arg(Arg::with_name("file")
                            .short("f")
                            .long("file")
                            .value_name("FILE")
                            .required(true)
                            .help("TOML file specifying the run context.")
                            .takes_value(true)).get_matches();

    let config_contents = load_config(args.value_of("FILE").unwrap());
    let config_contents = match config_contents {
        Err(err) => {
            println!("Error reading config file: {}", err);
            std::process::exit(1);
        },
        Ok(c) => c,
    };

    let parsed_config = parse_config(&config_contents);

    match parsed_config {
        Ok(config) => println!("Running {}", config.run),
        Err(err) => {
            println!("Error parsing config: {}", err);
            std::process::exit(1);
        },
    };
}

// Parse a TOML formatted specification file
fn parse_config(cfg: &Vec<u8>) -> Result<Config, de::Error> {
    let config = toml::from_slice(cfg);
    return config;
}

// Find and read a config file
fn load_config(file_path: &str) -> Result<Vec<u8>, io::Error> {
    let path = Path::new(file_path);
    let mut file = File::open(&path)?;
    let mut contents = Vec::new();
    file.read_to_end(&mut contents)?;
    return Ok(contents);
}
